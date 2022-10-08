1. 架构图
![image-20221006201035134](http://markdown-pic-june.oss-cn-beijing.aliyuncs.com/uPic/image-20221006201035134.png)

2. 功能介绍
- 广播消息
- 一对一私聊
- 改名
- 显示登录用户
- 上线提醒
- 不活跃强踢（代码设定的是5min）

3. 启动说明

首先启动服务器 server.go，其次启动客户端：

推荐使用nc命令登录客户端`nc 127.0.0.1 8788`，指令如下：
- `rename xxx`    改名
- `to zhangsan`   message 私聊
- `message`       直接输入消息就是广播消息
- `users`        显示当前在线用户

当然也可以使用代码中提供的client


4. 其他

- 代码注释很全，可以流畅阅读
- 代码经验吸取自 https://github.com/aceld/zinx