# goSearch
查询样例
```bash
curl --location --request POST '43.142.195.165:9999/api/v1/search' \
--header 'Content-Type: application/json' \
--data-raw '{
    "query":"今天星期几",
    "page":1,
    "limit":10,
    "filterWord":["搜索"]
}'
```
结果
```bash
{
    "time": 273,
    "total": 1000,
    "pageCount": 100,
    "page": 1,
    "limit": 10,
    "documents": [
        {
            "id": 2726222,
            "text": "我<span style=\"color: red;\">今天</span>才刚到报考,下<span style=\"color: red;\">星期</span>六就要考试了.我们都好好加油吧.",
            "url": "https://gimg2.baidu.com/image_search/src=http%3A%2F%2Fn.sinaimg.cn%2Ftranslate%2F552%2Fw600h752%2F20190104%2FAFtc-hrfcctm3432871.jpg&refer=http%3A%2F%2Fn.sinaimg.cn&app=2002&size=f9999,10000&q=a80&n=0&g=0n&fmt=jpeg?sec=1631721546&t=6bc772c3079216f9ffaac0501fed1b04",
            "score": 1.9451573860825984
        },
        {
            "id": 5199571,
            "text": "我上个<span style=\"color: red;\">星期</span>五做的手术,到<span style=\"color: red;\">今天</span>刚刚好一个<span style=\"color: red;\">星期</span>,做完",
            "url": "https://gimg2.baidu.com/image_search/src=http%3A%2F%2Fstatic4.j.cn%2Fimg%2Fforum%2F171229%2F1457%2F7739359341e246ed.jpg&refer=http%3A%2F%2Fstatic4.j.cn&app=2002&size=f9999,10000&q=a80&n=0&g=0n&fmt=jpeg?sec=1631748641&t=ab34ea20b956482ff25e1aa43a3de796",
            "score": 1.9451573860825984
        },
        ......
    ],
    "related": [
        "今天星期几"
    ],
    "words": [
        "星期",
        "今天"
    ]
}
```
