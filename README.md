# 手撕苟 —— 爬到墙外去，让我念句诗 (- -)

通过Web API动态创建“阵法”通道。

纯属个人学习苟语言的产物，三无产品无保障，欢迎并感谢star。

## 安装

暂无，描绘中……

## 接口设计

GET     /api/v1/ports => shows detail of all ports.

POST    /api/v1/ports [port, password, method] => create user port.

GET     /api/v1/details => shows server details. [bandwidth,traffic,load]

## 用法

$ curl /api/v1/ports

[
    {
        "port": 1080,
        "token": "example",
        "cipher": "aes-256-cfb"
    }
]

$ curl -X post --data '{"port":10080,"password":"example","method":"aes-256-cfb"}' /api/v1/ports

{
    "code": 200,
    "message": "OK",
    "data": null
}