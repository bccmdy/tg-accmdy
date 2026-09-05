# tg-accmdy
image
使用说明
1. 启动命令


bash
TELEGRAM_TOKEN=xxx \
GATEWAY_ENDPOINT=http://127.0.0.1:50066/v1 \
GATEWAY_API_KEY=your_bearer_token \
go run main.go

2. 添加提供商（支持自动模型）

markdown
/addprovider

回复：


markdown
name=Grok2API&type=gateway&endpoint=http://127.0.0.1:50066/v1&apikey=sk-xxx&model=   ← 留空或写 auto 即可自动获取模型

3. 测试命令

markdown
# 普通图片
/generate 一只穿着太空服的猫在月球上 2

# 透明背景图（PNG Alpha通道）
/transparent 一只透明猫在透明背景上 1

4. 透明图效果
返回的数据链接以 data:image/png;base64, 开头，Telegram 会直接显示高清透明 PNG
建议提示词中加入 transparent background、png with alpha channel 等关键词
