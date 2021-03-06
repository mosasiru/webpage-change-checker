# webpage-change-checker

Check webpages you like, and notify if it changed. 

<img src="https://user-images.githubusercontent.com/1458744/34559931-368f6306-f187-11e7-8e51-998f14be9dfd.png" width="50%">



# Usage

```bash
go get github.com/mosasiru/webpage-change-checker

cp $GOPATH/src/github.com/mosasiru/webpage-change-checker/config.examble.toml config.toml
```

write your own config, then


```bash
webpage-change-checker -c config.toml
```



# Configuration



## [[checker.pages]]

you can set multi pages.

|name            |type  |description                                 |default         |note                                |
|----------------|------|--------------------------------------------|----------------|------------------------------------|
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
|cache_dir            |string|the directory name where the latest response is saved  | .checker           |||


## [slack]


|name            |type  |description                                 |default         |note                                |
|----------------|------|--------------------------------------------|----------------|------------------------------------|
|webhook_url       |string|the url of webhook  |            | required |
|channel            |string|the channel name  |            ||
|username            |string|notifier name  |            ||
|icon_emoji            |string|notifier icon emoji  |            ||
|icon_url            |string|notifier icon url  |            ||
|alert_prefix            |string|the prefix used when any change or error occurs   |            |example: <!here>, <!channel>|

