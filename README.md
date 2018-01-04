# webpage-change-checker

Check webpages you like, and notify if it changed. 

<img src="https://user-images.githubusercontent.com/1458744/34559931-368f6306-f187-11e7-8e51-998f14be9dfd.png" width="50%">



# Usage

`cp config.example.toml config.toml`

write your own config, and

```
go get ./...
go run checker.go
```



# Configuration



## [[checker.pages]]

you can set multi pages.

|name            |type  |description                                 |default         |note                                |
|----------------|------|--------------------------------------------|----------------|------------------------------------|
|name            |string|site name  |            |required|
|url            |string|site url  |            |required|
|selector            |string|CSS selector you want to check diff  |            |example: #content|
|timeout            |int|request timeout (second)  |     10       ||
|interval            |int|check interval (second)  |     600       |don't attack sites!|
|notify_no_change            |bool|notify when nothing is changed also  |            ||
|notify_error            |bool|notify  when error occurs  |     10       |false|
|notifier            |enum("slack")| notifier. slack is only available now.  |           |||

## [checker]


|name            |type  |description                                 |default         |note                                |
|----------------|------|--------------------------------------------|----------------|------------------------------------|
|cache_file            |string|the file name prefix where pre response is saved  | .checker.           |||


## [slack]


|name            |type  |description                                 |default         |note                                |
|----------------|------|--------------------------------------------|----------------|------------------------------------|
|webhook_url       |string|the url of webhook  |            | required |
|channel            |string|the channel name  |            ||
|username            |string|notifier name  |            ||
|icon_emoji            |string|notifier icon emoji  |            ||
|icon_url            |string|notifier icon url  |            ||
|alert_prefix            |string|the prefix used when any change or error occurs   |            |example: <!here>, <!channel>|

