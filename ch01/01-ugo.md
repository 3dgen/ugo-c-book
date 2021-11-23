# µGo 简介
µGo 是迷你Go语言玩具版本，只保留最基本的int数据类型、变量定义和函数、分支和循环等最基本的特性。µGo 有以下的关键字：`var`、`func`、`if`、`for`、`return`。此外有一个内置的数据类型`int`，以及内置函数`func input() int`用于读取一个整数、内置函数`println(...)`用于打印输出。

比如计算1到100的和对应以下代码：

```go
func main() {
	var sum int
	for i := 0; i <= 100; i = i+1 {
		sum = sum+1
	}
	println(sum)
}
```

µGo 的具体定义会根据需要和实现的情况调整，目前可以将其作为Go的最小子集。