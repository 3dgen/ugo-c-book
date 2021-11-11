# ugo-c-book

本项目是柴大(github.com/chai2010)ugo编译器项目(github.com/chai2010/ugo-compiler-book)的镜像项目，与原项目输出LLVM汇编代码不同的是，本项目的输出为C代码。

# µGo 介绍
µGo 是迷你Go语言玩具版本，只保留最基本的int数据类型、变量定义和函数、分支和循环等最基本的特性。µGo 有以下的关键字：var、func、if、for、return。此外有一个int内置的数据类型，func input() int 函数读取一个整数，println(...) 打印函数。

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

µGo 的具体定义会根据需要和实现的情况调整，目前可以将其作为Go的最小子集就可。