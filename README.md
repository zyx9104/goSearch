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
