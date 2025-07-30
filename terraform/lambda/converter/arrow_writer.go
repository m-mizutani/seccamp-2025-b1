package main

import (
	"bytes"
	"fmt"

	"github.com/apache/arrow/go/v17/arrow"
	"github.com/apache/arrow/go/v17/arrow/array"
	"github.com/apache/arrow/go/v17/arrow/memory"
	"github.com/apache/arrow/go/v17/parquet"
	"github.com/apache/arrow/go/v17/parquet/compress"
	"github.com/apache/arrow/go/v17/parquet/pqarrow"
)

func buildOCSFSchema() *arrow.Schema {
	return arrow.NewSchema([]arrow.Field{
		{Name: "category_uid", Type: arrow.PrimitiveTypes.Int64},
		{Name: "class_uid", Type: arrow.PrimitiveTypes.Int64},
		{Name: "type_uid", Type: arrow.PrimitiveTypes.Int64},
		{Name: "activity_id", Type: arrow.PrimitiveTypes.Int64},
		{Name: "severity_id", Type: arrow.PrimitiveTypes.Int64},
		{Name: "time", Type: arrow.PrimitiveTypes.Int64},
		{Name: "start_time", Type: arrow.PrimitiveTypes.Int64, Nullable: true},
		{Name: "end_time", Type: arrow.PrimitiveTypes.Int64, Nullable: true},
		{Name: "status_id", Type: arrow.PrimitiveTypes.Int64},
		{Name: "confidence", Type: arrow.PrimitiveTypes.Int64, Nullable: true},
		{Name: "actor", Type: arrow.StructOf(
			arrow.Field{Name: "user", Type: arrow.StructOf(
				arrow.Field{Name: "name", Type: arrow.BinaryTypes.String},
				arrow.Field{Name: "uid", Type: arrow.BinaryTypes.String},
				arrow.Field{Name: "email_addr", Type: arrow.BinaryTypes.String},
				arrow.Field{Name: "domain", Type: arrow.BinaryTypes.String, Nullable: true},
				arrow.Field{Name: "type_id", Type: arrow.PrimitiveTypes.Int64},
				arrow.Field{Name: "groups", Type: arrow.ListOf(arrow.BinaryTypes.String), Nullable: true},
			)},
			arrow.Field{Name: "session", Type: arrow.StructOf(
				arrow.Field{Name: "uid", Type: arrow.BinaryTypes.String},
				arrow.Field{Name: "created_time", Type: arrow.PrimitiveTypes.Int64, Nullable: true},
				arrow.Field{Name: "exp_time", Type: arrow.PrimitiveTypes.Int64, Nullable: true},
			), Nullable: true},
			arrow.Field{Name: "app_name", Type: arrow.BinaryTypes.String, Nullable: true},
			arrow.Field{Name: "app_uid", Type: arrow.BinaryTypes.String, Nullable: true},
		)},
		{Name: "api", Type: arrow.StructOf(
			arrow.Field{Name: "service", Type: arrow.StructOf(
				arrow.Field{Name: "name", Type: arrow.BinaryTypes.String},
				arrow.Field{Name: "version", Type: arrow.BinaryTypes.String, Nullable: true},
			)},
			arrow.Field{Name: "operation", Type: arrow.BinaryTypes.String},
			arrow.Field{Name: "request", Type: arrow.StructOf(
				arrow.Field{Name: "uid", Type: arrow.BinaryTypes.String},
			)},
			arrow.Field{Name: "response", Type: arrow.StructOf(
				arrow.Field{Name: "code", Type: arrow.PrimitiveTypes.Int64},
				arrow.Field{Name: "message", Type: arrow.BinaryTypes.String, Nullable: true},
			), Nullable: true},
		)},
		{Name: "cloud", Type: arrow.StructOf(
			arrow.Field{Name: "provider", Type: arrow.BinaryTypes.String},
			arrow.Field{Name: "account", Type: arrow.StructOf(
				arrow.Field{Name: "uid", Type: arrow.BinaryTypes.String},
				arrow.Field{Name: "name", Type: arrow.BinaryTypes.String, Nullable: true},
			)},
			arrow.Field{Name: "org", Type: arrow.StructOf(
				arrow.Field{Name: "name", Type: arrow.BinaryTypes.String},
				arrow.Field{Name: "uid", Type: arrow.BinaryTypes.String, Nullable: true},
			), Nullable: true},
			arrow.Field{Name: "cloud_region", Type: arrow.BinaryTypes.String, Nullable: true},
		)},
		{Name: "src_endpoint", Type: arrow.StructOf(
			arrow.Field{Name: "ip", Type: arrow.BinaryTypes.String},
			arrow.Field{Name: "hostname", Type: arrow.BinaryTypes.String, Nullable: true},
			arrow.Field{Name: "location", Type: arrow.StructOf(
				arrow.Field{Name: "country", Type: arrow.BinaryTypes.String, Nullable: true},
				arrow.Field{Name: "src_region", Type: arrow.BinaryTypes.String, Nullable: true},
				arrow.Field{Name: "city", Type: arrow.BinaryTypes.String, Nullable: true},
			), Nullable: true},
		)},
		{Name: "web_resources", Type: arrow.ListOf(arrow.StructOf(
			arrow.Field{Name: "name", Type: arrow.BinaryTypes.String, Nullable: true},
			arrow.Field{Name: "uid", Type: arrow.BinaryTypes.String, Nullable: true},
			arrow.Field{Name: "type", Type: arrow.BinaryTypes.String, Nullable: true},
			arrow.Field{Name: "url_string", Type: arrow.BinaryTypes.String, Nullable: true},
			arrow.Field{Name: "data", Type: arrow.StructOf(
				arrow.Field{Name: "classification", Type: arrow.BinaryTypes.String, Nullable: true},
			), Nullable: true},
		)), Nullable: true},
		{Name: "metadata", Type: arrow.StructOf(
			arrow.Field{Name: "correlation_uid", Type: arrow.BinaryTypes.String, Nullable: true},
			arrow.Field{Name: "labels", Type: arrow.ListOf(arrow.BinaryTypes.String), Nullable: true},
			arrow.Field{Name: "original_time", Type: arrow.BinaryTypes.String, Nullable: true},
			arrow.Field{Name: "processed", Type: arrow.PrimitiveTypes.Int64, Nullable: true},
			arrow.Field{Name: "product_name", Type: arrow.BinaryTypes.String, Nullable: true},
			arrow.Field{Name: "version", Type: arrow.BinaryTypes.String, Nullable: true},
		), Nullable: true},
		{Name: "observables", Type: arrow.ListOf(arrow.StructOf(
			arrow.Field{Name: "name", Type: arrow.BinaryTypes.String},
			arrow.Field{Name: "type", Type: arrow.BinaryTypes.String},
			arrow.Field{Name: "value", Type: arrow.BinaryTypes.String},
		)), Nullable: true},
		{Name: "aws_region", Type: arrow.BinaryTypes.String},
		{Name: "account_id", Type: arrow.BinaryTypes.String},
		{Name: "event_hour", Type: arrow.BinaryTypes.String},
	}, nil)
}

func generateOCSFParquetFileArrow(logs []OCSFWebResourceActivity) ([]byte, error) {
	schema := buildOCSFSchema()
	mem := memory.NewGoAllocator()
	
	// Create record builder
	recordBuilder := array.NewRecordBuilder(mem, schema)
	defer recordBuilder.Release()

	// Build records
	for _, log := range logs {
		// Basic fields
		recordBuilder.Field(0).(*array.Int64Builder).Append(int64(log.CategoryUID))
		recordBuilder.Field(1).(*array.Int64Builder).Append(int64(log.ClassUID))
		recordBuilder.Field(2).(*array.Int64Builder).Append(int64(log.TypeUID))
		recordBuilder.Field(3).(*array.Int64Builder).Append(int64(log.ActivityID))
		recordBuilder.Field(4).(*array.Int64Builder).Append(int64(log.SeverityID))
		recordBuilder.Field(5).(*array.Int64Builder).Append(log.Time)
		
		// Optional fields
		if log.StartTime != 0 {
			recordBuilder.Field(6).(*array.Int64Builder).Append(log.StartTime)
		} else {
			recordBuilder.Field(6).(*array.Int64Builder).AppendNull()
		}
		
		if log.EndTime != 0 {
			recordBuilder.Field(7).(*array.Int64Builder).Append(log.EndTime)
		} else {
			recordBuilder.Field(7).(*array.Int64Builder).AppendNull()
		}
		
		recordBuilder.Field(8).(*array.Int64Builder).Append(int64(log.StatusID))
		
		if log.Confidence != 0 {
			recordBuilder.Field(9).(*array.Int64Builder).Append(int64(log.Confidence))
		} else {
			recordBuilder.Field(9).(*array.Int64Builder).AppendNull()
		}

		// Actor struct
		actorBuilder := recordBuilder.Field(10).(*array.StructBuilder)
		actorBuilder.Append(true)
		
		// Actor.User struct
		userBuilder := actorBuilder.FieldBuilder(0).(*array.StructBuilder)
		userBuilder.Append(true)
		userBuilder.FieldBuilder(0).(*array.StringBuilder).Append(log.Actor.User.Name)
		userBuilder.FieldBuilder(1).(*array.StringBuilder).Append(log.Actor.User.UID)
		userBuilder.FieldBuilder(2).(*array.StringBuilder).Append(log.Actor.User.EmailAddr)
		if log.Actor.User.Domain != "" {
			userBuilder.FieldBuilder(3).(*array.StringBuilder).Append(log.Actor.User.Domain)
		} else {
			userBuilder.FieldBuilder(3).(*array.StringBuilder).AppendNull()
		}
		userBuilder.FieldBuilder(4).(*array.Int64Builder).Append(int64(log.Actor.User.TypeID))
		
		// Groups (list)
		groupsBuilder := userBuilder.FieldBuilder(5).(*array.ListBuilder)
		if len(log.Actor.User.Groups) > 0 {
			groupsBuilder.Append(true)
			groupsValueBuilder := groupsBuilder.ValueBuilder().(*array.StringBuilder)
			for _, group := range log.Actor.User.Groups {
				groupsValueBuilder.Append(group)
			}
		} else {
			groupsBuilder.AppendNull()
		}
		
		// Actor.Session struct (nullable)
		sessionBuilder := actorBuilder.FieldBuilder(1).(*array.StructBuilder)
		if log.Actor.Session.UID != "" {
			sessionBuilder.Append(true)
			sessionBuilder.FieldBuilder(0).(*array.StringBuilder).Append(log.Actor.Session.UID)
			if log.Actor.Session.CreatedTime != 0 {
				sessionBuilder.FieldBuilder(1).(*array.Int64Builder).Append(log.Actor.Session.CreatedTime)
			} else {
				sessionBuilder.FieldBuilder(1).(*array.Int64Builder).AppendNull()
			}
			if log.Actor.Session.ExpTime != 0 {
				sessionBuilder.FieldBuilder(2).(*array.Int64Builder).Append(log.Actor.Session.ExpTime)
			} else {
				sessionBuilder.FieldBuilder(2).(*array.Int64Builder).AppendNull()
			}
		} else {
			sessionBuilder.AppendNull()
		}
		
		// Actor app fields
		if log.Actor.AppName != "" {
			actorBuilder.FieldBuilder(2).(*array.StringBuilder).Append(log.Actor.AppName)
		} else {
			actorBuilder.FieldBuilder(2).(*array.StringBuilder).AppendNull()
		}
		
		if log.Actor.AppUID != "" {
			actorBuilder.FieldBuilder(3).(*array.StringBuilder).Append(log.Actor.AppUID)
		} else {
			actorBuilder.FieldBuilder(3).(*array.StringBuilder).AppendNull()
		}

		// API struct
		apiBuilder := recordBuilder.Field(11).(*array.StructBuilder)
		apiBuilder.Append(true)
		
		// API.Service
		serviceBuilder := apiBuilder.FieldBuilder(0).(*array.StructBuilder)
		serviceBuilder.Append(true)
		serviceBuilder.FieldBuilder(0).(*array.StringBuilder).Append(log.API.Service.Name)
		if log.API.Service.Version != "" {
			serviceBuilder.FieldBuilder(1).(*array.StringBuilder).Append(log.API.Service.Version)
		} else {
			serviceBuilder.FieldBuilder(1).(*array.StringBuilder).AppendNull()
		}
		
		apiBuilder.FieldBuilder(1).(*array.StringBuilder).Append(log.API.Operation)
		
		// API.Request
		requestBuilder := apiBuilder.FieldBuilder(2).(*array.StructBuilder)
		requestBuilder.Append(true)
		requestBuilder.FieldBuilder(0).(*array.StringBuilder).Append(log.API.Request.UID)
		
		// API.Response (nullable)
		responseBuilder := apiBuilder.FieldBuilder(3).(*array.StructBuilder)
		if log.API.Response.Code != 0 {
			responseBuilder.Append(true)
			responseBuilder.FieldBuilder(0).(*array.Int64Builder).Append(int64(log.API.Response.Code))
			if log.API.Response.Message != "" {
				responseBuilder.FieldBuilder(1).(*array.StringBuilder).Append(log.API.Response.Message)
			} else {
				responseBuilder.FieldBuilder(1).(*array.StringBuilder).AppendNull()
			}
		} else {
			responseBuilder.AppendNull()
		}

		// Cloud struct
		cloudBuilder := recordBuilder.Field(12).(*array.StructBuilder)
		cloudBuilder.Append(true)
		cloudBuilder.FieldBuilder(0).(*array.StringBuilder).Append(log.Cloud.Provider)
		
		// Cloud.Account
		accountBuilder := cloudBuilder.FieldBuilder(1).(*array.StructBuilder)
		accountBuilder.Append(true)
		accountBuilder.FieldBuilder(0).(*array.StringBuilder).Append(log.Cloud.Account.UID)
		if log.Cloud.Account.Name != "" {
			accountBuilder.FieldBuilder(1).(*array.StringBuilder).Append(log.Cloud.Account.Name)
		} else {
			accountBuilder.FieldBuilder(1).(*array.StringBuilder).AppendNull()
		}
		
		// Cloud.Org (nullable)
		orgBuilder := cloudBuilder.FieldBuilder(2).(*array.StructBuilder)
		if log.Cloud.Org.Name != "" {
			orgBuilder.Append(true)
			orgBuilder.FieldBuilder(0).(*array.StringBuilder).Append(log.Cloud.Org.Name)
			if log.Cloud.Org.UID != "" {
				orgBuilder.FieldBuilder(1).(*array.StringBuilder).Append(log.Cloud.Org.UID)
			} else {
				orgBuilder.FieldBuilder(1).(*array.StringBuilder).AppendNull()
			}
		} else {
			orgBuilder.AppendNull()
		}
		
		// Cloud region
		if log.Cloud.Region != "" {
			cloudBuilder.FieldBuilder(3).(*array.StringBuilder).Append(log.Cloud.Region)
		} else {
			cloudBuilder.FieldBuilder(3).(*array.StringBuilder).AppendNull()
		}

		// SrcEndpoint struct
		srcEndpointBuilder := recordBuilder.Field(13).(*array.StructBuilder)
		srcEndpointBuilder.Append(true)
		srcEndpointBuilder.FieldBuilder(0).(*array.StringBuilder).Append(log.SrcEndpoint.IP)
		
		if log.SrcEndpoint.Hostname != "" {
			srcEndpointBuilder.FieldBuilder(1).(*array.StringBuilder).Append(log.SrcEndpoint.Hostname)
		} else {
			srcEndpointBuilder.FieldBuilder(1).(*array.StringBuilder).AppendNull()
		}
		
		// SrcEndpoint.Location (nullable)
		locationBuilder := srcEndpointBuilder.FieldBuilder(2).(*array.StructBuilder)
		if log.SrcEndpoint.Location.Country != "" || log.SrcEndpoint.Location.Region != "" || log.SrcEndpoint.Location.City != "" {
			locationBuilder.Append(true)
			if log.SrcEndpoint.Location.Country != "" {
				locationBuilder.FieldBuilder(0).(*array.StringBuilder).Append(log.SrcEndpoint.Location.Country)
			} else {
				locationBuilder.FieldBuilder(0).(*array.StringBuilder).AppendNull()
			}
			if log.SrcEndpoint.Location.Region != "" {
				locationBuilder.FieldBuilder(1).(*array.StringBuilder).Append(log.SrcEndpoint.Location.Region)
			} else {
				locationBuilder.FieldBuilder(1).(*array.StringBuilder).AppendNull()
			}
			if log.SrcEndpoint.Location.City != "" {
				locationBuilder.FieldBuilder(2).(*array.StringBuilder).Append(log.SrcEndpoint.Location.City)
			} else {
				locationBuilder.FieldBuilder(2).(*array.StringBuilder).AppendNull()
			}
		} else {
			locationBuilder.AppendNull()
		}

		// WebResources (list of structs)
		webResourcesBuilder := recordBuilder.Field(14).(*array.ListBuilder)
		if len(log.WebResources) > 0 {
			webResourcesBuilder.Append(true)
			webResourcesValueBuilder := webResourcesBuilder.ValueBuilder().(*array.StructBuilder)
			for _, resource := range log.WebResources {
				webResourcesValueBuilder.Append(true)
				if resource.Name != "" {
					webResourcesValueBuilder.FieldBuilder(0).(*array.StringBuilder).Append(resource.Name)
				} else {
					webResourcesValueBuilder.FieldBuilder(0).(*array.StringBuilder).AppendNull()
				}
				if resource.UID != "" {
					webResourcesValueBuilder.FieldBuilder(1).(*array.StringBuilder).Append(resource.UID)
				} else {
					webResourcesValueBuilder.FieldBuilder(1).(*array.StringBuilder).AppendNull()
				}
				if resource.Type != "" {
					webResourcesValueBuilder.FieldBuilder(2).(*array.StringBuilder).Append(resource.Type)
				} else {
					webResourcesValueBuilder.FieldBuilder(2).(*array.StringBuilder).AppendNull()
				}
				if resource.URLString != "" {
					webResourcesValueBuilder.FieldBuilder(3).(*array.StringBuilder).Append(resource.URLString)
				} else {
					webResourcesValueBuilder.FieldBuilder(3).(*array.StringBuilder).AppendNull()
				}
				
				// Data struct
				dataBuilder := webResourcesValueBuilder.FieldBuilder(4).(*array.StructBuilder)
				if resource.Data.Classification != "" {
					dataBuilder.Append(true)
					dataBuilder.FieldBuilder(0).(*array.StringBuilder).Append(resource.Data.Classification)
				} else {
					dataBuilder.AppendNull()
				}
			}
		} else {
			webResourcesBuilder.AppendNull()
		}

		// Metadata struct (nullable)
		metadataBuilder := recordBuilder.Field(15).(*array.StructBuilder)
		metadataBuilder.Append(true)
		
		if log.Metadata.CorrelationUID != "" {
			metadataBuilder.FieldBuilder(0).(*array.StringBuilder).Append(log.Metadata.CorrelationUID)
		} else {
			metadataBuilder.FieldBuilder(0).(*array.StringBuilder).AppendNull()
		}
		
		// Labels (list)
		labelsBuilder := metadataBuilder.FieldBuilder(1).(*array.ListBuilder)
		if len(log.Metadata.Labels) > 0 {
			labelsBuilder.Append(true)
			labelsValueBuilder := labelsBuilder.ValueBuilder().(*array.StringBuilder)
			for _, label := range log.Metadata.Labels {
				labelsValueBuilder.Append(label)
			}
		} else {
			labelsBuilder.AppendNull()
		}
		
		if log.Metadata.OriginalTime != "" {
			metadataBuilder.FieldBuilder(2).(*array.StringBuilder).Append(log.Metadata.OriginalTime)
		} else {
			metadataBuilder.FieldBuilder(2).(*array.StringBuilder).AppendNull()
		}
		
		if log.Metadata.Processed != 0 {
			metadataBuilder.FieldBuilder(3).(*array.Int64Builder).Append(log.Metadata.Processed)
		} else {
			metadataBuilder.FieldBuilder(3).(*array.Int64Builder).AppendNull()
		}
		
		if log.Metadata.ProductName != "" {
			metadataBuilder.FieldBuilder(4).(*array.StringBuilder).Append(log.Metadata.ProductName)
		} else {
			metadataBuilder.FieldBuilder(4).(*array.StringBuilder).AppendNull()
		}
		
		if log.Metadata.Version != "" {
			metadataBuilder.FieldBuilder(5).(*array.StringBuilder).Append(log.Metadata.Version)
		} else {
			metadataBuilder.FieldBuilder(5).(*array.StringBuilder).AppendNull()
		}

		// Observables (list of structs)
		observablesBuilder := recordBuilder.Field(16).(*array.ListBuilder)
		if len(log.Observables) > 0 {
			observablesBuilder.Append(true)
			observablesValueBuilder := observablesBuilder.ValueBuilder().(*array.StructBuilder)
			for _, observable := range log.Observables {
				observablesValueBuilder.Append(true)
				observablesValueBuilder.FieldBuilder(0).(*array.StringBuilder).Append(observable.Name)
				observablesValueBuilder.FieldBuilder(1).(*array.StringBuilder).Append(observable.Type)
				observablesValueBuilder.FieldBuilder(2).(*array.StringBuilder).Append(observable.Value)
			}
		} else {
			observablesBuilder.AppendNull()
		}

		// Simple string fields at the end
		recordBuilder.Field(17).(*array.StringBuilder).Append(log.Region)
		recordBuilder.Field(18).(*array.StringBuilder).Append(log.AccountID)
		recordBuilder.Field(19).(*array.StringBuilder).Append(log.EventHour)
	}

	// Build the record
	record := recordBuilder.NewRecord()
	defer record.Release()

	// Write to Parquet
	var buf bytes.Buffer
	props := parquet.NewWriterProperties(
		parquet.WithCompression(compress.Codecs.Snappy),
	)
	
	writer, err := pqarrow.NewFileWriter(schema, &buf, props, pqarrow.DefaultWriterProps())
	if err != nil {
		return nil, fmt.Errorf("failed to create parquet writer: %w", err)
	}
	defer writer.Close()

	if err := writer.Write(record); err != nil {
		return nil, fmt.Errorf("failed to write record: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	return buf.Bytes(), nil
}