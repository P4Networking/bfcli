# Protocol Independent Switch Control - Command Line Interface(PISC-CLI)

## Development envinment

- GVM(https://github.com/moovweb/gvm) 做 GO 版本管理
- GoLand 作為 IDE
- Golang version 1.14+
- Barefoot SDE version 9.2

## Function Overview
- completion : provide table name suggestion to user
- version : show version of PISC-CLI
- set-flow : insert entry into the table
- del-flow : delete entry/entries of the table
- table : list all the table name
- dump : print all entries of the table
- info : print table's information

## Run
PISC-CLI有Completion功能，補助使用者輸入Table名稱，因此執行方式有兩種。
1. 要使用completion功能，需要把PISC-CLI的路徑加到PATH裡頭
    ```
    cd WHERE/IS/PISC-CLI
    export PATH=$PATH:$PWD
    ```
    再把PISC-CLI的Completion功能引進BASH
    ```
    . <(pisc-cli completion)
    ```
    就可以使用 tab key 所提供Command與Table Name的推薦(At now, dump, info support only)
    ```
    pisc-cli [command] [flags] [arguments]
    ```
2. 若不想使用completion，直接執行PISC-CLI即可
    ```
    ./pisc-cli [command] [flags] [arguments]
    ```
    
## How to set a flow into device?
set-flow指令，因為table有三種不同的match方式，為了使用指令必須要按照規則，規則如下：
```
pisc-cli set-flow [TABLE NAME] [ACTION NAME] -m "match_key, ..." -a "action_value, ..."
```
且Match key和Action value的寫入順序必須要跟pipeline所設定的順序相同。


>設table key的時候請務必確認，你所設定的值是否跟table所需要的值是一致的，假如說table需要MAC Addr，但卻你下的指令是沒有格式的數字或IP, 這樣會直接設下去，且會產生問題。

1. Exact : smac table example. 
    ```
    pisc-cli set-flow pipe.SwitchIngress.smac SwitchIngress.smac_hit -m "aa:aa:aa:aa:aa:aa" -a "10"
    ```
    
2. LPM : rib table example。對於LPM, 必須要按照CIDR的形式輸入，否則會出現error.
    ```
    pisc-cli set-flow pipe.SwitchIngress.rib SwitchIngress.hit_route -m "10.0.0.1/24" -a "10"
    pisc-cli set-flow rib SwitchIngress.hit_route -m "10.0.0.1/24" -a "10" 
    ```

3. Ternary : acl example。對於Ternary，value與mask的長度以及形式必須要相同。Ternary match key的區分僅能靠Match key的priority的欄位。PISC-CLI利用 match flag來題共設定priority。priority的設置方式與其他match key相同，且順序也是位於match的最後面。

    ```
    pisc-cli set-flow pipe.SwitchIngress.arp SwitchIngress.arp_response -m "511/511, 0xffff, ffff, 255.255.255.255/255.255.255.255, 4000(priority value)" -a "ff:ff:ff:ff:ff:ff"    pisc-cli set-flow acl SwitchIngress.output -m "10/10, 0x0800/0xffff, fa:aa:aa:aa:aa:fa/ff:ff:ff:ff:ff:ff, 127.0.0.1/255.255.255.255, 11.1.1.2/255.255.255.255" -a "0"
    ```

## TTL - table TIMEOUT value setup
SMAC table has an idle_timeout function. This function will send a notification to the controller to notify that the entry's life has expired and need to remove it.
In PISC-CLI, it delivers TTL setting by -t flag on set-flow function. 
The scale of the flag is milliseconds. min, max value of the TTL need to refer PISC.

Example as following:

    ```
    pisc-cli set-flow pipe.SwitchIngress.smac SwitchIngress.smac_hit -m "ff:ff:ff:ff:ff:ff" -a "511" -t 60000
    pisc-cli set-flow smac SwitchIngress.smac_hit -m "ff:ff:ff:ff:ff:ff" -a "511" -t 60000
    ```

## How to delete entry/entries from device?
```
pisc-cli del-flow [TABLE_NAME] [Flag] [Arguments]
```
1. reset : clear all tables
    ```
    pisc-cli del-flow -r
    ```
2. all : delete all entries of the table
    ```
    // Do not confuse with set-flow's "-a" flag. It's totally diffrent.
    pisc-cli del-flow [TABLE_NAME] -a
    pisc-cli del-flow pipe.SwitchIngress.fib -a
    pisc-cli del-flow fib -a
    ```
3. match : delete specific entry by match keys.
    the match flag acts like set-flow function.
    the match flag need to specify table's match keys to delete entry.
    ```
    pisc-cli del-flow [TABLE_NAME] -m "Match key, ..."
    pisc-cli del-flow pipe.SwitchIngress.rib -m "192.168.1.7/255.255.255.255, 4000"
    pisc-cli del-flow rib -m "192.168.1.7/255.255.255.255, 4000"
    ```

## How to show the table and entries information?
```
pisc-cli [command] [TABLE_NAME]
```
1. table : list all table name.
    ```
    pisc-cli table
    ```
2. dump :  dump all entries of the table. 
    ```
    pisc-cli dump [TABLE_NAME]
    pisc-cli dump pipe.SwitchIngress.fib or pisc-cli dump fib
    ```
    ```
    // dump all tables entries
    pisc-cli dump -a
    // dump all entries counts of the table.
    pisc-cli dump -c 
    // also can dump all tables entries counts.
    pisc-cli dump -a -c
    ```
3. info : show table information to the user
    ```
    pisc-cli info [TALBE_NAME]
    ```

## Show version and logo
1. show version
    ```
    pisc-cli version
    ```

2. show logo of PISC
    ```
    pisc-cli version -l
    ```