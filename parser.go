package main

import (
    "fmt"
    "bytes"
    "os/exec"
    "strings"
    "encoding/json"
    "time"
    "reflect"
    "strconv"
)

func GetCmdInfo(c, arg string) string {
        var out bytes.Buffer
        var stderr bytes.Buffer
        var cmdout string
        done := make(chan error, 1)

        parts := strings.Fields(c)
        head  := parts[0]
        parts  = parts[1:len(parts)]
        parts = append(parts, arg)

        cmd := exec.Command(head, parts...)
        cmd.Stdout = &out
        cmd.Stderr = &stderr
        err := cmd.Start()
        if err != nil {
                fmt.Println("cmd error: ", err)
        }

        go func() {
            done <- cmd.Wait()
        }()

        select {
            case <- time.After(5 * time.Second):
                cmd.Process.Kill()
                cmdout = "timeout"
            case err := <-done:
                if err == nil {
                    cmdout = out.String()
                }
        }

        return cmdout
}

func getValue(data string) string {
    return strings.Trim(strings.Split(data, "=")[1], "\"")
}

func getTag(data string) string {
    return getValue(strings.Split(data, ",")[1])
}

/*------------------------nic basic--------------------------*/
type NicBasic struct {
    Bandwidth string `json:"bandwidth"`
    Mac       string `json:"mac"`
    Slot      string `json:"slot"`
    Manufacturer string `json:"manufacturer"`
    Model        string `json:"model"`      
}

type NicOutPut struct {
    state string
    data  map[string]NicBasic
}

func NicResCmp (outputv2, outputv3 interface{}) string {
    o2, ok2 := outputv2.(NicOutPut)
    o3, ok3 := outputv3.(NicOutPut)

    if ok2 && ok3 { 
        if "timeout" == o2.state || "timeout" == o3.state {
            return "timeout"
        } else {
            return strconv.FormatBool(reflect.DeepEqual(o2, o3)) 
        }
    } 
    fmt.Println("nic data trans failed")

    return "false"
}

func getNicBasicInfo(data string) NicBasic {
    var basic NicBasic
    values := strings.Split(data, ",")
    for _, v := range values {
        switch {
            case strings.Contains(v, "bandwidth"):
                basic.Bandwidth = getValue(v)
            case strings.Contains(v, "mac"):
                basic.Mac = getValue(v)
            case strings.Contains(v, "slot"):
                basic.Slot = getValue(v)
            case strings.Contains(v, "manufacturer"):
                basic.Manufacturer = getValue(v)
            case strings.Contains(v, "model"):
                basic.Model = getValue(v)
        }
    }

    return basic
}
func NicV3Output(cmd string, arg string) interface{} {
    var output NicOutPut
    output.data = make(map[string]NicBasic)

    cmdout := GetCmdInfo(cmd, arg)
    if "timeout" == cmdout {
        output.state = "timeout"
        return output 
    }

    output.state = "ok"

    items := strings.Split(cmdout, "\n")
    for _, item := range items {
        item_arr := strings.SplitAfterN(item, " ", 2)
        if 0 != len(item) {
            output.data[strings.TrimSpace(getTag(item_arr[0]))] = getNicBasicInfo(item_arr[1])
        }
    }    
    return output
}

type NicItem struct {
    Basic      NicBasic `json:"basic"`
    Macro_name string `json:"macro_name"`
}
func NicV2Output(cmd, arg string) interface{} {
    //cmdout := `[{"basic":{"bandwidth":"1Gbit/s","create_time":"1506620384","duplex":"Full","mac":"00:0c:29:4d:60:e7","manufacturer":"Intel Corporation","model":"82545EM Gigabit Ethernet Controller (Copper)","name":"ens33","slot":"0","version":"None"},"macro_name":"ens33"}]`

    cmdout := GetCmdInfo(cmd, arg)

    var output NicOutPut
    output.state = "ok"

    var items []NicItem
    err := json.Unmarshal([]byte(cmdout), &items)
    if err != nil {
        fmt.Println("error: ", err)
    }

    output.data = make(map[string]NicBasic)
    for _, item := range items {
        output.data[item.Macro_name] = item.Basic 
    }

    return output
}

/*-----------------------os basic---------------------------*/
type OsBasic struct {
    Kernel     string `json:"kernel"`
    Version    string `json:"version"`
}
type OsItem struct {
    Basic      OsBasic `json:"basic"`
}

type OsOutPut struct {
    state string
    data  OsBasic
}

func OsResCmp (outputv2, outputv3 interface{}) string {
    o2, ok2 := outputv2.(OsOutPut)
    o3, ok3 := outputv3.(OsOutPut)

    if ok2 && ok3 { 
        if "timeout" == o2.state || "timeout" == o3.state {
            return "timeout"
        } else {
            return strconv.FormatBool(reflect.DeepEqual(o2, o3)) 
        }
    } 
    
    fmt.Println("os data trans failed")
    
    return "false"
}

func OsV2Output(cmd, arg string) interface{} {
   /*cmdout :=    `[
      {"basic":{"create_time":"1506961596","kernel":"3.10.0-514.16.1.el7.x86_64","version":"CentOS Linux 7.3"}}
]`*/

    cmdout := GetCmdInfo(cmd, arg)

    var output OsOutPut
    output.state = "ok"

    var items []OsItem
    err := json.Unmarshal([]byte(cmdout), &items)
    if err != nil {
        fmt.Println("error: ", err)
    }

    output.data = items[0].Basic
    time.Sleep(1 * time.Second)

    return output

}

func getOsBasic(data string) OsBasic {
    var basic OsBasic
    values := strings.Split(data, ",")
    for _, v := range values {
        switch {
            case strings.Contains(v, "kernel"):
                basic.Kernel = getValue(v)
            case strings.Contains(v, "version"):
                basic.Version = getValue(v)
        }
    }

    return basic
}

func OsV3Output(cmd, arg string) interface{} {
    var output OsOutPut
   
    cmdout := GetCmdInfo(cmd, arg)
    if "timeout" == cmdout {
        output.state = "timeout"
        return output
    }

    output.state = "ok"

    data := strings.Trim(cmdout, "\n")
    item_arr := strings.SplitAfterN(data, " ", 2)
    output.data = getOsBasic(item_arr[1])

    return output 
}

/*-----------------------CPU--------------------------*/
func CpuResCmp (outputv2, outputv3 interface{}) string {
    return "timeout"
}

func CpuV2Output(cmd, arg string) interface{} {
    var output NicOutPut
    output.data = make(map[string]NicBasic)

    cmdout := GetCmdInfo(cmd, arg)
    if "timeout" == cmdout {
        output.state = "timeout"
        return output 
    }

    output.state = "ok"

    items := strings.Split(cmdout, "\n")
    for _, item := range items {
        item_arr := strings.SplitAfterN(item, " ", 2)
        if 0 != len(item) {
            output.data[strings.TrimSpace(getTag(item_arr[0]))] = getNicBasicInfo(item_arr[1])
        }
    }    
    return output
}

func CpuV3Output(cmd, arg string) interface{} {
    var output NicOutPut
    output.data = make(map[string]NicBasic)
    nicbasic := NicBasic{}

    output.data["cpu1_v3"] = nicbasic
    time.Sleep(1 * time.Second)

    return output
}

/*-----------------------DIMM--------------------------*/
type DimmBasic struct {
    Cap       string `json:"capacity"`
    LogSlot   string `json:"logical_slot"`
    Manu      string `json:"manufacturer"`
    MaxFreq   string `json:"max_freq"`
    Model     string `json:"model"`      
    PhySlot   string `json:"physical_slot"`      
    Sn        string `json:"sn"`      
    Volt      string `json:"volt"`      
}

type DimmItem struct {
    Basic      DimmBasic `json:"basic"`
    Macro_name string `json:"macro_name"`
}

type DimmOutPut struct {
    state string
    data  map[string]DimmBasic
}

func DimmResCmp (outputv2, outputv3 interface{}) string {
    o2, ok2 := outputv2.(DimmOutPut)
    o3, ok3 := outputv3.(DimmOutPut)

    if ok2 && ok3 { 
        if "timeout" == o2.state || "timeout" == o3.state {
            return "timeout"
        } else {
            return strconv.FormatBool(reflect.DeepEqual(o2, o3)) 
        }
    } 
    fmt.Println("dimm basic data trans failed")

    return "false"
}

func DimmV2Output(cmd, arg string) interface{} {
cmdout := `[
  {"basic":{"capacity":"1024","create_time":"1507529156","logical_slot":"RAM slot #0","manufacturer":"NONE","max_freq":"0","model":"NONE","physical_slot":"RAM slot #0","sn":"NONE_0","volt":"0"},"macro_name":"dimm0","performance":{"cur_freq":"0","numa":"OFF"}}
]
`
    var output DimmOutPut
    output.state = "ok"

    var items []DimmItem
    err := json.Unmarshal([]byte(cmdout), &items)
    if err != nil {
        fmt.Println("error: ", err)
    }

    output.data = make(map[string]DimmBasic)
    for _, item := range items {
        output.data[item.Macro_name] = item.Basic 
    }

    time.Sleep(1 * time.Second)

    return output
}

func DimmV3Output(cmd, arg string) interface{} {
    var output DimmOutPut
    output.data = make(map[string]DimmBasic)
    dimmbasic := DimmBasic{}

    output.data["dimm1_v3"] = dimmbasic
    time.Sleep(1 * time.Second)

    output.state = "ok"

    return output
}


/*
func main() {
    fmt.Printf("test agent start......\n\n")

   GetV3Output("nic.basic", "./nic_basic", "")

   GetV2Output("nic.basic", cmdout)

   fmt.Printf("\ntest agent end......\n")
}
*/
