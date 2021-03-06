
# 基于端口转发的非想天则观战的实验，研究与实现

如今，网络地址转换（Network Address Translation，NAT）设备如光猫、路由器等日益普及，许多联网设备失去了被从公网访问的能力。尽管某些情况下可以通过调整配置解决，例如对所有途径的NAT设备进行端口映射如配置虚拟服务器、DMZ主机，在多数情况下人们并没有调整配置的权限，如小区宽带，运营商NAT，网吧路由器，甚至拿不到超级帐号的光猫等等。

类毛玉正是为了解决这个问题而出现的。有别于虚拟局域网如Hamachi、QQ对战平台、浩方对战平台、游侠对战平台，类毛玉通过将非想天则主机的端口转发到公网服务器上，使其它用户可以通过公网服务器的端口连接到处于内网的主机，从而进行对战。

## 类毛玉的历史

* 毛玉，2014年前，由[zh99998](https://github.com/zh99998)开发，目标是一个二次元游戏平台，后来重心转移出则圈。
* Garden，2014年初，在毛玉进入则圈时，因不满基于NW.js的毛玉的大小，使用C#实现的类毛玉服务，将大小缩小为毛玉的百分之一，并支持记录对战结果和导出到天则观数据库的功能，仅在小范围内使用。
* 黑毛玉，2015年前后，使用毛玉源代码架设的类毛玉服务，由黑红进行运营，后合并入新毛玉。
* SokuHook，2015年末，因刚接触Win32Hook打算练手，使用C实现的类毛玉服务，启动后注入到非想天则中，企图降低延迟并共用非想天则主机的端口。
* SokuMKII，2016年初，因SokuHook在WinXP下兼容性不好，基于NW.js重新实现的类毛玉服务，加入了延迟展示和启动游戏的功能。
* Phantom，2016年末，因SokuMKII代码年久失修，运用新近学习的知识，基于Electron实现的类毛玉服务，加入了使用WebRTC进行NAT穿透的试验性功能。
* 新毛玉，2017年中，毛玉改版后回归，与萌卡结合紧密。
* Shitama，2017年中，因SokuMKII和Phantom多多少少存在问题，使用Go和Qt/C++重新实现的类毛玉服务，加入了支持观战的试验性功能。

## 非想天则的联机机制

建立主机时会监听一个UDP端口，默认是「10800」。
客机在连入之前会监听两个UDP端口，主端口默认是「clientport」，为「0」时由系统选择，副端口随机选择，用于测试NAT环境。
客机主端口和副端口向主机端口发送「0x1」数据包，主机端口接收后回复「0x3」数据包，用于测试连接能否建立，假如接收不到「0x3」数据包，客机提示「连接失败」。
剩余的数据将只通过主机端口和客机主端口进行传输。

主机与客机之间只使用一个端口进行数据传输，因此为非想天则实现端口转发可以是简单甚至简陋的。
在知识储备不足时，为了支持端口限制型NAT和对称型NAT，SokuMKII和Phantom都只支持单端口转发，首个连接将被绑定，无数据传输数秒后解除绑定，才能接受新的连接。
在编写Shitama的过程中，从Go的[yamux](https://github.com/hashicorp/yamux)实现多路传输中得到了启发，Shitama为每个转发的数据包标记了来源地址，解除了连接数的限制，复用同一个中转地址不再需要等待或多次连接了。

## 非想天则的观战机制

观战一直都是很强的需求点，然而正如上面所说，直到Shitama才真正支持多端口转发。虽然端口转发方面没有了障碍，中转地址的观战仍然是失败的。
于是对非想天则观战机制的数据测试就开始了。使用Wireshark可以很方便地捕获传输的数据包，寻找它们之间的规律，显然这是一个苦力活。

### 本机观战测试

使用Sandboxie，你可以在一台电脑上同时启动多个非想天则客户端。使用2P连入1P进行对战，3P连入1P进行观战，根据Wireshark捕获到的数据包，可以总结如下：
3P连入1P进行观战时，1P会告知3P「2P的地址」，3P使用这个地址与2P进行连接，并从2P获取观战数据。

### 外网观战测试

通过随机观战并捕获数据，可以总结如下：
上述结论成立，并且当2P处于端口限制型NAT或对称型NAT后时，3P不能进行观战。

因此，假如能够让1P直接发送观战数据给3P，而不是让3P去找2P，是否就能实现中转观战了呢？
于是开始使用IDA Pro分析非想天则的网络部分的代码，这又是一个苦力活，虽然对比数据包后的理解更加深刻。
以及尝试联系Fireseal希望能够得到一些指点，结果得知SokuRoll的源代码已经失传。

### 大规模外网观战测试

就在认定上述结论为真理时，在兔群观测到了一次复杂的观战事件，刷新了我对观战机制的认识。
当连接1P时，被重定向到3P，观战成功。再次连接1P时被重定向到2P。当连接2P时，被重定向到1-1P。并且在1-1P断开连接后，重新询问2P，被重定向到3P，观战得以继续。这意味着：
每一个非想天则客户端都维护了一个其他客户端的列表，重定向的目标客户端是随机的可变的。

### 本机中转观战测试

因此，如果我们保证维护的每个客户端地址是正确的，不是中转地址或者回环地址，客户端之间互相知道对方的地址，是有可能连上的。
当我在本机使用中转并尝试进行观战时，观战并没有成功，但却没有提示「连接失败」，捕获的数据包表明客户端在不断尝试新的端口，并在一段时间后，询问「是否观战」。
于是当我在「2017年6月19日2时30分」将「0x1」，「0x2」与「0x8」数据包中的地址都修正以后，本机中转观战变得可用，外网中转观战（对称型NAT）失败，外网中转观战（2P有公网IP）成功。

非想天则本身有一定的UDP穿透策略，在多年前可能比较好用，但近年来NAT环境日益复杂，感觉即使Shitama对此提供支持，效果也不会太明显。

## 非想天则的数据包结构

待整理。
