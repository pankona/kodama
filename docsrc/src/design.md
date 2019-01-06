# 非同期ジョブ実行システム設計

## ユースケース

## モジュール構成図

![](/assets/module-diagram.svg)

## シーケンス図

```uml
@startuml

skinparam monochrome true
skinparam defaultFontSize 20
skinparam defaultFontName courier

hide footbox

participant "**Web App**"    as app
participant "**Worker**"     as worker
box "Async Job System"
participant "**API Server**" as api
participant "**Job Queue**"  as queue
participant "**DB**"         as db
end box

== Register Job ==

app -> api : Register
    api -> queue : Register
        queue -> db : Insert\nnew Job
        queue <-- db : Job ID
    api <-- queue : Job ID
app <-- api : Job ID

== Execute Job ==

api <- worker : Fetch
    api -> queue : Fetch
        queue -> db : Select\nPending Job
        queue <-- db : Pending Job
    api <-- queue : Pending Job
api --> worker : Pending Job

worker -> worker : Execute Job
note right of worker : It may take long time...

api <- worker : Update (success/failure)
    api -> queue : update
        note right of queue: if job executing was failure\nand retry count doesn't\nexceed as configured, change\njob status to "Pending"
        queue -> db : Update\nJob status
        queue <-- db
    api <-- queue
api --> worker

== Confirm Job Status ==

app -> api : Status
    api -> queue : Status
        queue -> db : Select\nSpecified Job
        queue <-- db : Job
    api <-- queue : Job
app <-- api : Job

@enduml
```
