---
sidebar_position: 5
---

# HTTP API

## 端点

| 端点 | 方法 | 说明 |
|------|------|------|
| `/add` | POST | 添加延迟任务 |
| `/pop` | POST | 拉取就绪任务（长轮询） |
| `/finish` | POST | 确认完成 |
| `/cancel` | POST | 取消任务 |
| `/get` | GET | 查询任务 |
| `/stats` | GET | 队列统计 |

## 响应格式

```json
{"code": 0, "message": "ok", "data": null}
```

## HTTP 状态码

| 状态码 | 场景 |
|--------|------|
| 200 | 成功 |
| 400 | 参数错误 |
| 404 | 任务不存在 |
| 409 | 重复任务或 topic 冲突 |
| 503 | 队列已关闭 |
