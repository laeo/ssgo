# ssgo - a new shadowsocks server that controlled with web API.

credentials[port,token,cipher] 管理器监视凭据信息的增删操作，自动调用socket连接管理器创建、关闭net监听。



## Installation

### standalone !!! 1st

./ssgo --port=5001 --token=123456 --locally

### slave

./ssgo --ctrl-with=https://example.org/api/v1/slaves/i37e/auth?token=123456

## API Design


GET     /api/v1/users => shows detail of all users.

POST    /api/v1/users [port, token, method] => create user port.

GET     /api/v1/details => shows server details. [bandwidth,traffic,load]


## Usage

$ curl /api/v1/users

{
    [
        "port": 1080,
        "token": "example",
        "method": "aes-256-cfb"
    ]
}

$ curl -X post --data '{"port":10080,"token":"example","method":"aes-256-cfb"}' /api/v1/users

{
    "code": 200,
    "message": "OK",
    "data": null
}