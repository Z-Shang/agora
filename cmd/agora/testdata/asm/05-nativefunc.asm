//
// import "fmt" // implicit fmt variable
// fmt.Println("Hello ", "world")
//
[f]
05-nativefunc.agora
4
0
1
0
1
[k]
sfmt
sPrintln
sHello 
sworld
[i]
LOAD K 0 // <-fmt
POP V 0 // ->fmt
PUSH K 2 // <-"Hello "
PUSH K 3 // <-"world"
PUSH K 1 // <-"Println"
PUSH V 0 // <-"fmt"
CFLD A 2
PUSH N 0 // <-Nil
DUMP S 3
RET _ 0