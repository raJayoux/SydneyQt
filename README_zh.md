# SydneyQt

![SydneyQt](https://socialify.git.ci/juzeon/SydneyQt/image?font=Inter&forks=1&logo=https%3A%2F%2Fupload.wikimedia.org%2Fwikipedia%2Fcommons%2F9%2F9c%2FBing_Fluent_Logo.svg&name=1&owner=1&pattern=Signal&stargazers=1&theme=Light)

![Static Badge](https://img.shields.io/badge/project-SydneyQt-blue) ![GitHub release (with filter)](https://img.shields.io/github/v/release/juzeon/SydneyQt) ![GitHub all releases](https://img.shields.io/github/downloads/juzeon/SydneyQt/total) ![GitHub Workflow Status (with event)](https://img.shields.io/github/actions/workflow/status/juzeon/SydneyQt/wails.yml) ![GitHub License](https://img.shields.io/github/license/juzeon/SydneyQt)

[![SydneyQt - A desktop client for the jailbroken New Bing AI | Product Hunt](https://api.producthunt.com/widgets/embed-image/v1/featured.svg?post_id=438079&theme=light)](https://www.producthunt.com/posts/sydneyqt?utm_source=badge-featured&utm_medium=badge&utm_souce=badge-sydneyqt)

一个使用Go和[Wails](https://github.com/wailsapp/wails)构建的跨平台桌面客户端（[之前](https://github.com/juzeon/SydneyQt/tree/v1)基于Python和Qt），用于越狱新版Bing AI Copilot（Sydney版）。

## 特点

- 通过参数调整和提示注入越狱新版Bing。
- 提前使用灰度测试中的功能。
- 通过本地Selenium浏览器或远程Bypass Server自动解决CAPTCHA验证码。
- 使用代理和Cloudflare Workers解锁地区限制。
- 自由编辑聊天上下文，包括AI的之前的回复。
- 阻止Bing AI撤回消息，并自动发送自定义文本继续生成。
- 撤回和编辑你的最后一条消息。
- 制作、选择和发送自定义的快速回复到聊天中。
- 显示聊天上下文的富文本或纯文本，支持LaTeX公式、表格、代码等。
- 与你浏览的网页聊天。
- 与你打开的文件聊天（包括pdf、docx、pptx、xlsx和其他纯文本/代码文件）。
- Youtube视频总结。
- 具有视觉功能的GPT-4，支持图片搜索。
- 使用最新的 DALL·E 3 模型生成图像。
- 使用Bing的Suno模型生成音乐和音乐视频。
- 使用可切换的不同配置的OpenAI ChatGPT API。
- 在自定义的提示预设之间切换。
- 使用现代的Web技术构建的负责任和人性化的UI设计。
- 暗黑模式。
- 根据你的喜好自定义设置。

## 下载

你可以从[发布页面](https://github.com/juzeon/SydneyQt/releases)下载Windows、Linux和macOS的二进制文件，或者根据构建部分自己构建。

平台信息：

- Windows:  SydneyQt-windows-amd64.exe
- Linux:  SydneyQt-linux-amd64
- macOS: SydneyQt.app.zip, SydneyQt.pkg（未签名）

## 使用

1. 把你的`cookies.json`放在可执行文件的同一个文件夹中（对于macOS：`$HOME/Library/Application Support/SydneyQt`）：
   - 为[Chrome](https://chrome.google.com/webstore/detail/cookie-editor/hlkenndednhfkekhgcdicdfddnkalmdm)或[Firefox](https://addons.mozilla.org/en-US/firefox/addon/cookie-editor/)安装Cookie-Editor扩展（建议使用Chrome而不是Firefox，因为我们使用Chrome的网络栈来绕过Bing的防火墙和验证码）
   - 访问`bing.com`
   - 打开扩展
   - 授予所有网站的权限
   - 点击右下角的`Export`，然后点击`Export as JSON`（这会把你的cookies保存到剪贴板）
   - 把你的cookies粘贴到一个名为`cookies.json`的文件中，创建在可执行文件的同一个目录下。
   - **注意：在导出cookie之前，确保你可以使用网页聊天。**
2. 运行程序。

请按照下一节的说明解决常见问题。

## 常见问题

### 代理

对于中国大陆的用户，设置代理是必须的。

1. 在设置中尝试不同的代理类型。例如：http://127.0.0.1:7890, socks5://127.0.0.1:7890（假设7890是你的代理的端口）。
2. 如果你使用Clash或类似的代理软件，请确保带有`bing.com`后缀的域名通过代理路由。有些代理提供商可能把`bing.com`添加到直连规则中，这意味着它会绕过代理。
3. 如果这样也不行，把代理留空，并尝试使用[Proxifier](https://www.proxifier.com/)或Clash TUN模式。

### 地区污染

*仅限中国用户。*

如果你第一次在没有代理的情况下打开Bing网站，它会重定向你到`cn.bing.com`并污染你的cookies，这意味着你将无法再用这些cookies访问Bing AI，即使你之后使用了代理。如果发生地区污染，请先配置代理规则，确保Bing会通过代理访问，然后清除你浏览器的所有cookies或者打开一个隐私浏览窗口，重新登录你的Microsoft账号，最后导出cookies。

### Wss反向代理

Bing禁止特定国家访问Bing AI（具体来说，是sydney.bing.com），所以在这种情况下，你需要使用Cloudflare Workers设置一个wss反向代理。以下是操作步骤：

<details>
<summary>点击我</summary>

1. 访问[这个链接](https://dash.cloudflare.com/)并登录或注册一个Cloudflare账号。
2. 在侧边栏中，选择`Workers & Pages`。
3. 在打开的页面中，点击`Create application`。
4. 选择`Create Worker`。
5. 给你的worker一个名字，然后点击`Deploy`。
6. 在worker详情页面中，点击`Quick edit`。
7. 从[这里](https://raw.githubusercontent.com/juzeon/SydneyQt/v2/worker.js)复制所有的代码，然后粘贴到`worker.js`中的现有代码上。然后点击`Save and deploy`。
8. 复制worker域名，看起来像`xxxx-xxxx-xxxx.xxxx.workers.dev`（不是一个URL，像`https://xxxx-xxxx-xxxx.xxxx.workers.dev/`，请去掉前缀和后缀）并把它作为`Wss Domain`粘贴到设置页面中。然后点击`Save`。
</details>

### Cookie过期

你之前设置的cookies可能会不时过期。你可以在软件的聊天页面中检查你的cookies的状态。如果过期了，就按照使用部分中的cookies导入步骤重新操作。

### 验证码

从v2.4.0开始，SydneyQt将启动本地Selenium浏览器尝试自动解决验证码，并在配置了的情况下使用[Bypass服务器](https://github.com/Harry-zklcdc/go-proxy-bingai#%E4%BA%BA%E6%9C%BA%E9%AA%8C%E8%AF%81%E6%9C%8D%E5%8A%A1%E5%99%A8)。

如果不起作用，请按照以下步骤：

1. 检查cookies是否过期。如果是的话，重新导入它们。
2. 在确保cookies有效后，在你的浏览器中打开Bing Web并发送一个随机消息。你应该看到一个验证码挑战。如果没有，验证当前用户是否与cookies.json文件匹配。完成验证码后，回到软件。它应该可以正常工作了。

如果你遇到**无限验证码循环**，你可以尝试以下步骤：

1. 在你的手机上安装Bing移动版。

2. 使用你的Microsoft账号登录。

3. 向新版Bing发送一条消息。

**确保你的代理IP不会改变。**如果你使用Clash，请禁用负载均衡或轮询模式，只使用一个节点。否则你将需要在你的浏览器中频繁地手动解决验证码。

## 构建

环境：Go 1.21+，Node.js 16+

你可以按照 [Wails](https://wails.io/docs/gettingstarted/installation/) 的开发指南进行操作。

这里是简要版：

1. 安装 Go 和 Node.js。
2. 安装 Wails: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`。
3. 克隆项目: `git clone https://github.com/juzeon/SydneyQt`。
4. 运行构建命令: `wails build`。

### Developer Notes

使用文件`debug_options_sets.json`覆写optionsSets，例：

```json
[		
	"fluxsydney",
	"iyxapbing",
	"iycapbing",
	"clgalileoall",
	"gencontentv3",
	"nojbf"
]
```

## Web API

感谢 [@PeronGH](https://github.com/PeronGH) 现在我们有了一个 Web API。[点这里查看详情。](webapi/README.md)

## 截图

![](https://public.ptree.top/ShareX/2024/03/04/1709523428/sgKblqUfnA.png)

![](https://public.ptree.top/ShareX/2023/12/05/1701779864/syd-color.jpg)

![](https://public.ptree.top/ShareX/2024/03/04/1709523634/GbdrCn4VJf.png)

![](https://public.ptree.top/ShareX/2024/03/04/1709523618/lqXNgbfCt7.png)

![](https://public.ptree.top/ShareX/2024/03/04/1709523689/QQ3UWHSWpf.png)

![](https://public.ptree.top/ShareX/2024/03/04/1709523461/yGHPtEnYj3.png)

![](https://public.ptree.top/ShareX/2024/03/04/1709523494/dXybgQt0gg.png)

![](https://public.ptree.top/ShareX/2024/03/04/1709523476/6d3GqIwjV3.png)

![](https://public.ptree.top/ShareX/2024/03/04/1709523522/KWaHW1IfhU.png)

![](https://public.ptree.top/ShareX/2024/03/04/1709523563/BFTw4tcXdM.png)

![](https://public.ptree.top/ShareX/2024/03/04/1709523534/8NkIBWd8Yl.png)

![](https://public.ptree.top/ShareX/2024/03/04/1709523551/DYA9NyNXQP.png)



## Star 历史

[![Star History Chart](https://api.star-history.com/svg?repos=juzeon/SydneyQt&type=Date)](https://star-history.com/#juzeon/SydneyQt&Date)

## 致谢

<https://github.com/acheong08/EdgeGPT>

<https://github.com/InterestingDarkness/EdgeGPT/tree/sydney>

<https://github.com/Harry-zklcdc/go-proxy-bingai>