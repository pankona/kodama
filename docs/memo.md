# メモ

未分類文書は一旦ここに書く。

## 設計で考えた点

### 概要

- ウェブアプリケーションから用いるという文脈で、様々な非同期ジョブ実行を行うことができる仕組みを考える。

  - 色々なユースケースに対応するため、なるべく特定用途に寄らない設計にしたいと考える。
  - 構成は、ジョブはいったんキューに積み、順次ワーカーが処理を行っていくモデル

- 使う側からしてみたらシンプルに使える API を提供する

  - ジョブ投入
  - ジョブ取り出し
  - ジョブ状態確認
  - ジョブ状態変更

- 非同期ジョブ実行システムの基盤ということで、ジョブがそれなりに時間が掛かる (一日とか) でも使えるような想定

- ジョブ実行者 (ワーカーと呼ぶ) は、必要に応じてどんどん増やせる (スケールアウト) できるように

  - キュー部分は一元管理したい関係上、スケールアウトしにくいので、
    キューとワーカーは別にスケールできるようにしておく

- キューはストレージに永続化する (DB を使うだろう) のでスケールアウトしにくいという想定

### ジョブキュー (名前微妙)

- 投入されたジョブをキューイングしておく機能

  - 投入されたジョブはワーカーが順次取り出して実行する

- 投入されたジョブはストレージに置くなどして永続化する

  - ジョブキューのプロセスが死んだときに、投入済のジョブが失われないようにするため
  - ジョブキューをスケールアウトしやすい

- ジョブは以下の状態を持つ

  - 実行待ち

    - ワーカーによって処理されるのを待っている状態のジョブ

  - 実行中

    - ワーカーが実行中のジョブ
    - 一定期間が過ぎると「実行待ち」に戻る

  - 正常終了
  - 失敗終了
    - エラー理由を添えることができる

- キューといいつつ、取り出された段階で内容が失われるわけではない

  - ジョブが失敗したときに再実行される可能性があるので、一度取り出されても失われないようにする
  - ジョブがワーカーに取り出されたら、そのジョブは「実行中」に状態を変更する

    - 複数のワーカーが同じジョブを実行しないようにするため
    - 「実行中」状態は、一定の時間が過ぎると解除されて「実行待ち」に戻る
      - ワーカーが結果を通知せずに死んでしまったり、結果通知時に通信ができなかったり、などで結果が行方不明になる場合にそなえて
      - この仕様のため、「同一のジョブは二回以上実行されうる」という制限が生じる。

  - ワーカーがジョブを正常に終了した場合、ワーカーがジョブキューに「実行終了」を通知する

    - ジョブは「正常終了」に状態変更される

- ワーカーがジョブの実行を失敗する場合があることを想定する。

  - ジョブキューはワーカーがジョブの実行に失敗した回数を記録する
  - 失敗の回数は、リトライを行うかどうかの判断に用いる

    - リトライは 5 回行うとする
    - ジョブが失敗した場合、かつリトライ回数が 5 回に満たない場合、当該ジョブの状態を「実行待ち」に変更する

      - いずれワーカーが本ジョブをリトライする

    - ジョブが失敗した場合、かつリトライ回数が 5 回に達していた場合、当該ジョブの状態を「失敗終了」に変更する

      - システム管理者にメールを送信する

### ワーカー

- ジョブキューを見張っていて、ジョブが投入されていれば順次取りだして実行する機能
- 見張り方はポーリングかプッシュ型でも (gRPC とかロングポーリングとか WebSocket とか)

  - ジョブ投入から実行開始までに多少ディレイがあっても良い想定
  - 一日一回のポーリングとか。もっと頻度が高ければプッシュ側のほうが遊び時間が少なくて良くなりそう

- 実行が終わったら、結果をジョブキューに返す

  - ジョブの ID を引数にしてジョブの結果通知 API を実行する

### 結果通知

- システム側からのプッシュではなく利用側からのポーリングによって行うことを想定
- 一日単位などの長い処理が行われる場合、セッション張りっぱなしよりもポーリングのほうが仕組みが単純で安定すると思ったため
- ポーリングのほうがリソース面では不利 (コネクション張ったり切ったり) と思われるので、
  処理が短い場合はシステム側からのプッシュでも良さそう
- ジョブ投入時にジョブキューはジョブの ID を返す
  - 結果の問い合わせは、利用側はジョブ ID を引数にして現在のジョブの状態を問い合わせる形
