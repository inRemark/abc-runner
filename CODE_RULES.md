# CODE_RULES

## 文件描述信息

本项目为人类程序员和AI智能体（Qoder）共同实现，在文件顶部应该包含文件描述信息，最重要的是作者、创建日期、修改日期和文件描述。
因此AI(Qoder)创建文件时，在文件顶部新增作者信息，格式如下：

```go
// @Description: 文件描述信息
// @Since: {version}
// @Author: {Qoder}
// @Date: 2020-05-05 09:09:09
```

## 代码函数描述信息

函数描述信息应该包含函数功能、参数、返回值、作者、创建日期、修改日期和函数描述。

```go
// @Description: 
// @Since: {version}
// @Author: {Qoder}
// @Date: {yyyy}-{mm}-{dd}
func functionName(param1, param2 int) (int, error) {}
```
