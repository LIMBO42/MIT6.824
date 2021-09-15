```go
import (
	"fmt"
	"math"
)

//声明在变量后面
func add(x int, y int) int {
	return x + y
}

//var 语句用于声明一个变量列表，跟函数的参数列表一样，类型在最后。
var i, j int = 1, 2
//var可以省略类型，如果没有二义性的话
var c, python, java = true, false, "no!"

//在函数中，简洁赋值语句 := 可在类型明确的地方代替 var 声明。
//函数外的每个语句都必须以关键字开始（var, func 等等），因此 := 结构不能在函数外使用。
c, python, java := true, false, "no!"

//数值转换
var f float64=float64(i)
f:=float64(i)


//类型推导
//在声明一个变量而不指定其类型时（即使用不带类型的 := 语法或 var = 表达式语法），变量的类型由右值推导得出。当右值声明了类型时，新变量的类型与其相同

//Go 的 for 语句后面的三个构成部分外没有小括号， 大括号 { } 则是必须的。
for i:=0;i<10;i++{
    
}

//for是go里面的while
for sum<100{
    
}

//如果省略循环条件 就无限循环
for{
    
}

//if 与 for 类似，不需要括号
//可以写个赋值语句
if v := math.Pow(x, n); v < lim {
	return v
} else {
	fmt.Printf("%g >= %g\n", v, lim)
}

//Go只运行选定的case，且switch

//没有条件的switch-case
//其实就是if else
switch {
	case t.Hour() < 12:
		fmt.Println("Good morning!")
	case t.Hour() < 17:
		fmt.Println("Good afternoon.")
	default:
		fmt.Println("Good evening.")
}

//defer 语句会将函数推迟到外层函数返回之后执行。
//推迟的函数调用会被压入一个栈中。当外层函数返回时，被推迟的函数会按照后进先出的顺序调用。
//Defer通常用于简化执行各种清理操作的函数。
for i := 0; i < 10; i++ {
	defer fmt.Println(i)
}

//结构体
type Vertex struct {
	X int
	Y int
}
//这里的指针依旧用点访问，隐式
func main() {
	v := Vertex{1, 2}
	p := &v
	p.X = 1e9
	fmt.Println(v)
}

//数组 两种方式声明数组
var a [2]string
primes := [6]int{2, 3, 5, 7, 11, 13}

//slice类似于数组的引用（修改就会对底层进行修改）
/*
切片并不存储任何数据，它只是描述了底层数组中的一段。
更改切片的元素会修改其底层数组中对应的元素。
与它共享底层数组的切片都会观测到这些修改。
*/

//slice可以类似于python一样省略
//len(s)获取长度 cap(s)获取容量
//容量是指从切片的第一个元素到底层数组的最后一个元素的长度

s := []int{2, 3, 5, 7, 11, 13}
printSlice(s)

//len=0 cap=6
s = s[:0]
printSlice(s)

//len=4 cap=6
s = s[:4]
printSlice(s)

//len=2 cap=4
//注意这里的len=2，因为是对s[:4]再进行切片
s = s[2:]
printSlice(s)

//切片的零值是 nil。nil 切片的长度和容量为 0 且没有底层数组。
var s []int

//make,第二个参数为len，第三个参数为cap
b := make([]int, 0, 5)
printSlice("b", b)

//cap=5
c := b[:2]
printSlice("c", c)

//append 向切片添加元素，当切片底层的数组太小时，会返回一个新分配的数组

//range 返回两个值 一个为下标，一个为下标对应的元素
var pow = []int{1, 2, 4, 8, 16, 32, 64, 128}
for i, v := range pow {
	fmt.Printf("2**%d = %d\n", i, v)
}
//可以用_,v或者i,_来忽略


//=为赋值运算符，:=为定义和初始化
//所以=要和var一起使用，而:=只能在函数中使用，只能定义局部变量

//映射
type Vertex struct {
	Lat, Long float64
}
//相当于map[string]vertex为类型
var m map[string]Vertex
m=make(map[string]Vertex)
m[""]=vertex{}
或者
var m = map[string]Vertex{
	"Bell Labs": Vertex{
		40.68433, -74.39967,
	}
}

//获取删除 检测某个值是否存在
m[key] = elem
delete(m, key)
//ok返回值是否存在
elem, ok = m[key]

//函数也是值，可以像其他值一样传递
func compute(fn func(float64, float64) float64) float64 {
	return fn(3, 4)
}

//闭包
//一个匿名函数
//1 3 6 ....
func adder() func(int) int {
	sum := 0
	return func(x int) int {
		sum += x
		return sum
	}
}

func main() {
	pos, neg := adder(), adder()
	for i := 0; i < 10; i++ {
		fmt.Println(
			pos(i),
			neg(-2*i),
		)
	}
}

//方法就是一类带特殊的 接收者 参数的函数。
//方法接收者在它自己的参数列表内，位于 func 关键字和方法名之间。
//所以v vertex为接收者，调用用v.abs()
func (v Vertex) Abs() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

//方法只是个带接收者参数的函数。
//功能不变化
//接收者的类型定义和方法声明必须在同一包内；不能为内建类型声明方法。
//即不能为int8，float64之类声明方法
//type myFloat float64这样将float64改为myfloat类型是可以的

//对于某类型 T，接收者的类型可以用 *T 的文法。
//指针接收者的方法可以修改接收者指向的值（就像 Scale 在这做的）。由于方法经常需要修改它的接收者，指针接收者比值接收者更常用。
//这样定义的方法可以修改vertex的值
func (v *Vertex) Scale(f float64) {
	v.X = v.X * f
	v.Y = v.Y * f
}

//带指针参数的函数必须接受一个指针
//以指针为接收者的方法被调用时，接收者既能为值又能为指针
//v.Scale(5)，即便 v 是个值而非指针，带指针接收者的方法也能被直接调用。 也就是说，由于 Scale 方法有一个指针接收者，为方便起见，Go 会将语句 v.Scale(5) 解释为 (&v).Scale(5)
指针的优点
//首先，方法能够修改其接收者指向的值。
//其次，这样可以避免在每次调用方法时复制该值。若值的类型为大型结构体时，这样做会更加高效。

//接口类型：一组方法名定义的集合
type Abser interface {
	Abs() float64
}
//接口类型的变量可以保存任何实现了这些方法的值。
type Abser interface {
	Abs() float64
}
var a Abser
f := MyFloat(-math.Sqrt2)
v := Vertex{3, 4}

a = f  // a MyFloat 实现了 Abser
a = &v // a *Vertex 实现了 Abser


//类型通过实现一个接口的所有方法来实现该接口。
type I interface {
	M()
}

type T struct {
	S string
}

// 此方法表示类型 T 实现了接口 I，但我们无需显式声明此事。
func (t T) M() {
	fmt.Println(t.S)
}

//接口变量应该接受的值是实现了接口的类型的变量
//nil 接口值既不保存值也不保存具体类型。
//为 nil 接口调用方法会产生运行时错误，因为接口的元组内并未包含能够指明该调用哪个 具体 方法的类型。

//指定了零个方法的接口值被称为 *空接口：*

interface{}

var i interface{}
func describe(i interface{}) {
	fmt.Printf("(%v, %T)\n", i, i)
}

//该语句断言i保存了类型T，并返回底层的值给t
t := i.(T)
t,ok:=i.(T)

//类型选择
switch v := i.(type) {
case T:
    // v 的类型为 T
case S:
    // v 的类型为 S
default:
    // 没有匹配，v 与 i 的类型相同
}


//fmt包中定义了stringer接口
type Stringer interface {
    String() string
}
//这个就是对于person的实现
func (p Person) String() string {
	return fmt.Sprintf("%v (%v years)", p.Name, p.Age)
}
//注意这里的return 用的是sprintf


```

![image-20210911174637730](C:\Users\LIMBO\AppData\Roaming\Typora\typora-user-images\image-20210911174637730.png)

注意这里是重写了error()

![image-20210911182036579](C:\Users\LIMBO\AppData\Roaming\Typora\typora-user-images\image-20210911182036579.png)

```go

// func (T) Read(b []byte) (n int, err error)
// 接受切片
r := strings.NewReader("Hello, Reader!")

b := make([]byte, 8)
for {
	n, err := r.Read(b)
	fmt.Printf("n = %v err = %v b = %v\n", n, err, b)
	fmt.Printf("b[:n] = %q\n", b[:n])
	if err == io.EOF {
		break
	}
}

// goroutine
go f(x, y, z)
//函数或者方法前面加上go，启动新的go程执行



// 信道：带有类型的管道 使用<-来发送或接收值
func sum(s []int, c chan int) {
	sum := 0
	for _, v := range s {
		sum += v
	}
	c <- sum // 将和送入 c
}

func main() {
	s := []int{7, 2, 8, -9, 4, 0}

	c := make(chan int)
	go sum(s[:len(s)/2], c)
	go sum(s[len(s)/2:], c)
	x, y := <-c, <-c // 从 c 中接收

	fmt.Println(x, y, x+y)
}
/*
当创建一个Go协程时，创建这个Go协程的语句立即返回。与函数不同，程序流程不会等待Go协程结束再继续执行。程序流程在开启Go协程后立即返回并开始执行下一行代码，忽略Go协程的任何返回值。
在主协程存在时才能运行其他协程，主协程终止则程序终止，其他协程也将终止。
*/

https://blog.csdn.net/u011304970/article/details/76168257


```

**<img src="C:\Users\LIMBO\AppData\Roaming\Typora\typora-user-images\image-20210912100138043.png" alt="image-20210912100138043" style="zoom:67%;" />**

`select` 语句使一个 Go 程可以等待多个通信操作。

`select` 会阻塞到某个分支可以继续执行为止，这时就会执行该分支。当多个分支都准备好时会随机选择一个执行。

当 `select` 中的其它分支都没有准备好时，`default` 分支就会执行。<img src="C:\Users\LIMBO\AppData\Roaming\Typora\typora-user-images\image-20210912101959284.png" alt="image-20210912101959284" style="zoom: 67%;" />

![image-20210912102808484](C:\Users\LIMBO\AppData\Roaming\Typora\typora-user-images\image-20210912102808484.png)

[go练习：Web 爬虫_u014472777的专栏-CSDN博客](https://blog.csdn.net/u014472777/article/details/105166833)

