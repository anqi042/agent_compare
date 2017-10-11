package main

import (
    "fmt"
    "time"
    "github.com/fatih/color"
)

type GetFormattedFunc func(cmd, arg string) interface{}
type ResCompareFunc   func(outputv2, outputv3 interface{}) string

type Command struct {
    version  string
    cmd      string
    arg      string
    getdata  GetFormattedFunc
    ch       chan interface{}
}

type TestSuite struct {
    Name   string
    V2Cmd  Command
    V3Cmd  Command
    cmp    ResCompareFunc
}

const (
         NICBASICSQL = "select a.'basic.name' macro_name, a.'timestamp' 'basic.create_time', a.'basic.slot', a.'basic.manufacturer', a.'basic.version', a.'basic.model', a.'basic.name', a.'basic.mac', a.'basic.bandwidth' from interface_details a;"
         OSBASICSQL = "select a.'name' || ' ' || a.'major' || '.' || a.'minor' as 'basic.version', a.'timestamp' as 'basic.create_time', b.'version' as 'basic.kernel' from os_version a, kernel_info b;"
)

var testSuites = []TestSuite{
           {"nic", 
                 Command{"agent_v2", "./sysqueryi --json", NICBASICSQL, NicV2Output, outputv2}, 
                 Command{"agent_v3", "./nic_basic", "", NicV3Output, outputv3},
                 NicResCmp},
           {"cpu", 
                 Command{"agent_v2", "./test.sh", "select cpu", CpuV2Output, outputv2}, 
                 Command{"agent_v3", "./cpubasic", "-cpu", CpuV3Output, outputv3},
                 CpuResCmp},
           {"dimm", 
                 Command{"agent_v2", "./sysqueryi --json", "select dimm", DimmV2Output, outputv2}, 
                 Command{"agent_v3", "./dimmbasic", "-dimm", DimmV3Output, outputv3},
                 DimmResCmp},
           {"os", 
                 Command{"agent_v2", "./sysqueryi --json", OSBASICSQL, OsV2Output, outputv2}, 
                 Command{"agent_v3", "./os_basic", "", OsV3Output, outputv3},
                 OsResCmp},
}

func execCmd(c chan interface{}, cmd Command) {
        c <- cmd.getdata(cmd.cmd, cmd.arg)
}

var outputv2 chan interface{}
var outputv3 chan interface{}

type TemplateData struct {
    state string
    data  interface{}
}

func SuiteRun(c_res chan string, suite TestSuite) {
    suite.V2Cmd.ch = make(chan interface{})
    suite.V3Cmd.ch = make(chan interface{})

    go execCmd(suite.V2Cmd.ch, suite.V2Cmd)
    go execCmd(suite.V3Cmd.ch, suite.V3Cmd)

    fmt.Println("TestSuite:------------------> ", suite.Name)

    outputv2, outputv3 := <-suite.V2Cmd.ch, <-suite.V3Cmd.ch

    fmt.Println("  V2 output: ", outputv2)
    fmt.Println("  V3 output: ", outputv3)

    c_res <- suite.cmp(outputv2, outputv3)
}

func Compare(quit chan int) {
    suiteRes := make([]SuiteRes, 0)
     
    cmp_res_chan := make(chan string)

    for _, suite := range testSuites {
        go SuiteRun(cmp_res_chan, suite)

        cmd_res := <- cmp_res_chan
        suiteRes = append(suiteRes, SuiteRes{suite.Name, cmd_res})
    }
    
    ShowRes(suiteRes)

    quit <- 0
}

type SuiteRes struct {
    SuiteName string
    Res       string
}

func ShowRes(res []SuiteRes) {
    var suc  int
    var fail int
    var timeout int

    fmt.Println("")
    fmt.Println("______________________")
    fmt.Printf("|%-10s|%-10s|\n", "suites", "result")
    fmt.Println("——————————————————————")
    for _, r := range res {
        if r.Res == "true" {
            color.Green("|%-10s|%-10s|\n", r.SuiteName, r.Res)
            suc++
        } else if r.Res == "timeout" {
            color.Blue("|%-10s|%-10s|\n", r.SuiteName, r.Res)
            timeout++
        } else if r.Res == "false" {
            color.Red("|%-10s|%-10s|\n", r.SuiteName, r.Res)
            fail++
        }
    }
    fmt.Println("")

    fmt.Println("Successed suites: ", suc)
    fmt.Println("Failed    suites: ", fail)
    fmt.Println("Timeout   suites: ", timeout)
    fmt.Println("")
}


func ResetTimer(s time.Duration) <-chan time.Time {
    return time.After(s * time.Second)
}


var quit chan int

func main() {
    quit = make(chan int)
    go Compare(quit)

    <-quit
}
