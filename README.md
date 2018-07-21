# ReceiverDemo
demo for receiver server.

### 认证流程

一、一共涉及 3 个部分。

(1) 设备（device）。

(2) 服务器（receiver server）。

(3) 认证服务器（token server）。

二、认证过程

(1) 设备向认证服务器发送请求，请求分配一个新的 token。

(2) 认证服务器返回一个 token 给设备。

(3) 设备连接到服务器（receiver server），发送 Auth 认证消息给服务器（receiver server），在 Auth 认证消息中带有 “设备ID” 和 “token”。

(4) 服务器（receiver server）收到 Auth 消息之后，把 “设备ID” 和 “token” 发送到认证服务器去验证是否合法。

(5) 认证服务器把认证的结果返回给服务器（receiver server）。

(6) 服务器（receiver server）如果收到认证成功，则向设备发送认证成功消息，并且可以开始接受设备上传的事件数据。

(7) 服务器（receiver server）如果收到认证失败，则向设备发送认证失败消息，同时断开和设备的连接。
