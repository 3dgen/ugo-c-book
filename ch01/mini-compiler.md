# 最小编译器

编译器的基本功能是将一种语言编写的代码转化为另一种语言的代码，前者称为源代码，后者称为目标代码。先从最简单的开始：我们假设源代码中只有一个整数，假定为x，与之对应的目标代码程序运行后返回值为x。

本项目设计的编译器将使用Go语言编写，编译目标为C代码。运行后返回指定整数的C代码形如：

```c
int main(){
    return 123;
}
```

上例的C程序运行后返回值应为123。若该程序保存于名为a.c的文件中，我们使用gcc编译并执行它，在Linux命令行下执行：

```
$ gcc a.c
$ ./a.out
$ echo $?
$ 123
```

> **info** `$?`记录了Linux命令行上一个程序执行后的返回值。后文Windows环境中的`%ERRORLEVEL%`作用类似。

在Windows命令行下执行：
```
> gcc a.c
> a.exe
> echo %ERRORLEVEL%
> 123
```

程序返回值确为123。

我们的最小编译器将实现**从源代码中读取一个整数x，编译为返回值为x的C代码**这一功能，编译器的完整Go代码如下：

```go
package main

import (
	"fmt"
	"os"
	"os/exec"
)

var code = "42"

func main() {
	compile(string(code))
}

func compile(code string) {
	output := fmt.Sprintf(tmpl, code)
	os.WriteFile("a.out.c", []byte(output), 0666)
	exec.Command("gcc", "a.out.c").Run()
}

const tmpl = `
int main(){
	return %v;
}
`
```

`compile`是该编译器的核心函数，它从源代码`code`中读取整数，按照`tmpl`模板格式化为C代码，存入名为a.out.c的文件中，并调用gcc对其进行编译。

使用以下命令行运行我们的编译器：

```
$ go run main.go
$ a.out
$ echo $?
$ 42
```

可见它如实的返回了我们在源代码`code`中设定的42，这样我们就实现了一个编译整数到C代码的最小编译器。

> **info** 虽然本项目编译器的编译目标为C代码，我们仍然需要使用gcc这样的C语言编译器将成果进一步编译为可执行程序，以验证生成的C代码的正确性。