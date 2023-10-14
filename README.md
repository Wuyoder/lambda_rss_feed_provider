# AWS Lambda in Golang for Provide Podcast's RSS Feed

- Trigger by S3 event
- Transform the `feed.json` to `feed.rss`
- Upload the template `feed.json` to S3 bucket and replace with the correct content
- This Lambda will parse the json file into rss format and upload to S3 bucket
- Ref: [RSS Provider / Who can help. Lambda!](https://ithelp.ithome.com.tw/articles/10337627)
- Ref: [RSS Provider / Just append, don't hesitate.](https://ithelp.ithome.com.tw/articles/10338225)
- (just for 2023 ithome 30 days challenge)
