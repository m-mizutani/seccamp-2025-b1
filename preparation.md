# 事前準備

## 環境構築

以下のソフトウェアをご自身のラップトップPCで使えるように準備してください。

- Go
  - version 1.22 以上
  - インストール: https://golang.org/doc/install
- git
  - OS標準のものが入っていれば設定不要
  - インストール: https://git-scm.com/downloads
- エディタ
  - 未導入 & 特にこだわりがなければ Visual Studio Code 推奨
    - インストール: https://code.visualstudio.com/download
  - Go, SQLを編集するので好みの拡張や設定を入れておくのを推奨します

## サービス関連設定

以下のサービス・アカウントを用意しておいてください。

- Eメールアドレス（プロバイダは自由）
  - アラート通知を閲覧するためのSlack Workspaceへ招待します
  - アラート確認用WebページのログインにもSlack SSOを利用します
- GitHub アカウント
  - CIが設定されているGitHubリポジトリに招待します

## 事前学習

特に課題形式などにはしませんが、事前に以下について取り組んでいただけると助かります。

### Go言語

- Tour of Go https://go.dev/tour を一通りやっておく
- 基本的には以下のような処理が自分で書けるようになっていると良いです
  - 複数の構造体を別の構造体に変換する
  - 構造体 ↔ json の変換ができる
  - 基本的な制御構文

### SQL基礎

実習でAthena（Presto/Trinoベース）を使ったログ分析を行います。以下のようなSQL構文を理解しているとスムーズに進められます。

- SELECT, WHERE, GROUP BY, ORDER BY
- COUNT, SUMなどの集約関数
- JOINの基本概念
- 時刻関数の基本（`date_format`, `from_unixtime`など）

Web上の無料資料としては（英語ですが）以下のようなところを参照してみてください。
- [SQLボルト](https://sqlbolt.com/) - インタラクティブなSQL学習サイト
- [W3Schools SQL Tutorial](https://www.w3schools.com/sql/) - 基本的なSQL構文の解説

### セキュリティ監視基礎

基本的に2つの資料は同じ内容を別角度から表現しているだけなので、どちらかをメインに見ていただき、もう片方を補完する形で見てもらうで構いません。講義だとこれらすべては解説する時間がないので、これらを見ていただいた前提で講義をします。

- [実践セキュリティ監視基盤構築](https://zenn.dev/mizutani/books/secmon-platform) の1、2章
- [セキュリティ・キャンプ2024 全国大会【専門】Bクラス：セキュリティ監視入門](https://mztn.notion.site/4a1b43b9101c4f669f32f805b2393206?pvs=74)

### AWSについて

今回は実習でAWSを利用するため、基礎的な概念やアーキテクチャなどを事前に理解しておいていただきたいです。具体的にはAWS認定試験の [AWS Certified Cloud Practitioner](https://aws.amazon.com/jp/certification/certified-cloud-practitioner/?ch=sec&sec=rmg&d=1) 相当の知識＋今回利用するS3、Athena、SNSがどういう役割を担うものなのかの概要を把握しておいてもらえると助かります。

学習にあたっては以下のコンテンツが無料で利用できます。（もちろん書籍などを利用するのも良いです）

- [AWS Skill Builder](https://skillbuilder.aws)
- [AWS Architecture Center](https://aws.amazon.com/jp/architecture/)

また、余裕があればAWSは無料枠があるので、自分でアカウントを作ってリソースを作成するなどしてみるのもおすすめです。ただし、作ったリソースを放置すると課金がかさみ無料枠を使い切ってしまうため、注意してください。
