package logcore

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

// バイナリフォーマットのマジックナンバーとバージョン
const (
	BinaryMagicNumber = 0x4C47534D // "LGSM" (Log Generator Seed Magic)
	BinaryFormatVersion = 1
)

// バイナリヘッダー構造
type BinaryHeader struct {
	Magic     uint32    // マジックナンバー
	Version   uint16    // フォーマットバージョン
	Reserved  uint16    // 予約領域
	Timestamp int64     // 生成時刻（Unix秒）
	SeedCount uint32    // シード数
	Checksum  uint32    // CRC32チェックサム
}

// DayTemplateをバイナリ形式で保存
func (dt *DayTemplate) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	
	// ヘッダー書き込み
	header := BinaryHeader{
		Magic:     BinaryMagicNumber,
		Version:   BinaryFormatVersion,
		Reserved:  0,
		Timestamp: dt.Metadata.Generated.Unix(),
		SeedCount: uint32(len(dt.LogSeeds)),
		Checksum:  0, // 後で計算
	}
	
	if err := binary.Write(&buf, binary.LittleEndian, header); err != nil {
		return nil, fmt.Errorf("failed to write header: %w", err)
	}
	
	// 日付文字列（固定長10バイト: "2024-08-12"）
	dateBytes := make([]byte, 10)
	copy(dateBytes, dt.Date[:min(10, len(dt.Date))])
	buf.Write(dateBytes)
	
	// LogSeedデータ（各10バイト）
	for _, seed := range dt.LogSeeds {
		// Timestamp: 4バイト（相対秒数、最大86400）
		if err := binary.Write(&buf, binary.LittleEndian, uint32(seed.Timestamp)); err != nil {
			return nil, fmt.Errorf("failed to write timestamp: %w", err)
		}
		
		// EventType, UserIndex, ResourceIdx, Pattern: 各1バイト
		buf.WriteByte(seed.EventType)
		buf.WriteByte(seed.UserIndex)
		buf.WriteByte(seed.ResourceIdx)
		buf.WriteByte(seed.Pattern)
		
		// Seed: 4バイト
		if err := binary.Write(&buf, binary.LittleEndian, seed.Seed); err != nil {
			return nil, fmt.Errorf("failed to write seed: %w", err)
		}
	}
	
	// メタデータ（簡略化）
	metaBytes := dt.encodeMetadata()
	if err := binary.Write(&buf, binary.LittleEndian, uint32(len(metaBytes))); err != nil {
		return nil, fmt.Errorf("failed to write metadata length: %w", err)
	}
	buf.Write(metaBytes)
	
	return buf.Bytes(), nil
}

// バイナリデータからDayTemplateを復元
func (dt *DayTemplate) UnmarshalBinary(data []byte) error {
	buf := bytes.NewReader(data)
	
	// ヘッダー読み込み
	var header BinaryHeader
	if err := binary.Read(buf, binary.LittleEndian, &header); err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}
	
	// マジックナンバー確認
	if header.Magic != BinaryMagicNumber {
		return fmt.Errorf("invalid magic number: 0x%x", header.Magic)
	}
	
	// バージョン確認
	if header.Version > BinaryFormatVersion {
		return fmt.Errorf("unsupported format version: %d", header.Version)
	}
	
	// 日付読み込み
	dateBytes := make([]byte, 10)
	if _, err := buf.Read(dateBytes); err != nil {
		return fmt.Errorf("failed to read date: %w", err)
	}
	dt.Date = string(bytes.TrimRight(dateBytes, "\x00"))
	
	// LogSeedデータ読み込み
	dt.LogSeeds = make([]LogSeed, header.SeedCount)
	for i := uint32(0); i < header.SeedCount; i++ {
		var timestamp uint32
		if err := binary.Read(buf, binary.LittleEndian, &timestamp); err != nil {
			return fmt.Errorf("failed to read timestamp at %d: %w", i, err)
		}
		
		eventType, err := buf.ReadByte()
		if err != nil {
			return fmt.Errorf("failed to read event type at %d: %w", i, err)
		}
		userIndex, err := buf.ReadByte()
		if err != nil {
			return fmt.Errorf("failed to read user index at %d: %w", i, err)
		}
		resourceIdx, err := buf.ReadByte()
		if err != nil {
			return fmt.Errorf("failed to read resource index at %d: %w", i, err)
		}
		pattern, err := buf.ReadByte()
		if err != nil {
			return fmt.Errorf("failed to read pattern at %d: %w", i, err)
		}
		
		var seed uint32
		if err := binary.Read(buf, binary.LittleEndian, &seed); err != nil {
			return fmt.Errorf("failed to read seed at %d: %w", i, err)
		}
		
		dt.LogSeeds[i] = LogSeed{
			Timestamp:   int64(timestamp),
			EventType:   eventType,
			UserIndex:   userIndex,
			ResourceIdx: resourceIdx,
			Pattern:     pattern,
			Seed:        seed,
		}
	}
	
	// メタデータ読み込み
	var metaLength uint32
	if err := binary.Read(buf, binary.LittleEndian, &metaLength); err != nil {
		return fmt.Errorf("failed to read metadata length: %w", err)
	}
	
	metaBytes := make([]byte, metaLength)
	if _, err := buf.Read(metaBytes); err != nil {
		return fmt.Errorf("failed to read metadata: %w", err)
	}
	
	if err := dt.decodeMetadata(metaBytes); err != nil {
		return fmt.Errorf("failed to decode metadata: %w", err)
	}
	
	// 生成時刻復元
	dt.Metadata.Generated = time.Unix(header.Timestamp, 0)
	
	return nil
}

// メタデータのエンコード（簡略化）
func (dt *DayTemplate) encodeMetadata() []byte {
	var buf bytes.Buffer
	
	// TotalLogs
	binary.Write(&buf, binary.LittleEndian, uint32(dt.Metadata.TotalLogs))
	
	// NormalRatio
	binary.Write(&buf, binary.LittleEndian, dt.Metadata.NormalRatio)
	
	// AnomalyStats（主要な5つのパターンのみ）
	patterns := []string{"example1", "example2", "example3", "time_anomaly", "volume_anomaly"}
	for _, pattern := range patterns {
		count := dt.Metadata.AnomalyStats[pattern]
		binary.Write(&buf, binary.LittleEndian, uint32(count))
	}
	
	return buf.Bytes()
}

// メタデータのデコード
func (dt *DayTemplate) decodeMetadata(data []byte) error {
	buf := bytes.NewReader(data)
	
	// TotalLogs
	var totalLogs uint32
	if err := binary.Read(buf, binary.LittleEndian, &totalLogs); err != nil {
		return err
	}
	
	// NormalRatio
	var normalRatio float64
	if err := binary.Read(buf, binary.LittleEndian, &normalRatio); err != nil {
		return err
	}
	
	// AnomalyStats
	dt.Metadata = SeedMeta{
		TotalLogs:         int(totalLogs),
		NormalRatio:       normalRatio,
		AnomalyStats:      make(map[string]int),
		LogCoreVersion:    LogCoreVersion,
		SeedFormatVersion: SeedFormatVersion,
	}
	
	patterns := []string{"example1", "example2", "example3", "time_anomaly", "volume_anomaly"}
	total := 0
	for _, pattern := range patterns {
		var count uint32
		if err := binary.Read(buf, binary.LittleEndian, &count); err != nil {
			return err
		}
		dt.Metadata.AnomalyStats[pattern] = int(count)
		total += int(count)
	}
	dt.Metadata.AnomalyStats["total"] = total
	
	return nil
}

// gzip圧縮付きでバイナリ保存
func (dt *DayTemplate) MarshalBinaryCompressed() ([]byte, error) {
	// バイナリ形式にエンコード
	binaryData, err := dt.MarshalBinary()
	if err != nil {
		return nil, err
	}
	
	// gzip圧縮
	var compressed bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressed)
	
	if _, err := gzipWriter.Write(binaryData); err != nil {
		gzipWriter.Close()
		return nil, fmt.Errorf("failed to compress data: %w", err)
	}
	
	if err := gzipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}
	
	return compressed.Bytes(), nil
}

// gzip展開付きでバイナリ読み込み
func (dt *DayTemplate) UnmarshalBinaryCompressed(data []byte) error {
	// gzip展開
	gzipReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()
	
	decompressed, err := io.ReadAll(gzipReader)
	if err != nil {
		return fmt.Errorf("failed to decompress data: %w", err)
	}
	
	// バイナリ形式からデコード
	return dt.UnmarshalBinary(decompressed)
}

// ヘルパー関数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}