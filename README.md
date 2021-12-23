# telegram-bot

一个没啥用的Telegram Bot

## 功能

- 反向搜图出处
  - SauceNAO
    - Reply一张图片并使用`/sauce`命令
    - 或者私聊直接发图给BOT
  - ascii2d
    - Reply一张图片并使用`/ascii2d`命令
    - ※会在SauceNAO搜索失败时自动启动
- 骰子
  - `/dice`可以投掷1次D6
  - `/dice 7d20`可以投掷7次D20
  - 以此类推最多支持同时投掷100个D10000
  - ※只投掷一次时投掷次数可不写

## Docker

```shell
docker run -d \
  -v /path/to/config.yaml:/app/config.yaml \
  --name telegram-bot \
  --restart unless-stopped \
  nagaseyami/telegram-bot:latest
```

如果你需要让机器人走代理，你可以

```shell
docker run -d \
  -v /path/to/config.yaml:/app/config.yaml \
  --name telegram-bot \
  --restart unless-stopped \
  -e HTTP_PROXY=http://yourproxy:8080
  -e HTTPS_PROXY=http://youproxy:8080
  nagaseyami/telegram-bot:latest
```

## 编译

```shell
go build main.go
```

## 配置文件

初次启动之后会自动在同目录下生成一个`config.yaml`文件

```yaml
# Debug模式，会记录更多log
DebugMode: false

# Telegram Bot Token
# https://core.telegram.org/bots#6-botfather
TelegramBotToken: ""

# BOT有时会删除一些没用的消息，比如「搜索失败」「功能未启用」等
# 这里填写删除消息前的等待时间
# 格式采用time.Duration
# https://pkg.go.dev/time#ParseDuration
DeleteMessageInterval: 5s

# SauceNAO
SaucenaoConfig:
  Enable: false
  # ApiKey可以去SauceNAO注册账号免费申请
  ApiKey: ""

# Ascii2d
Ascii2dConfig:
  Enable: false
  # Ascii2d必须上传图片搜索，所以搜索前会先把图片缓存
  # 此处填写图片缓存文件夹位置
  TempFolderPath: /temp

# 骰子
DiceConfig:
  Enable: false
```

编辑完成之后再次启动即可，如果你想指定配置文件位置，你可以在启动时添加`--config /path/to/config.yaml`指定配置文件位置
