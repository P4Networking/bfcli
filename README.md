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


1. Exact : smac table example
    ```
    pisc-cli set-flow pipe.SwitchIngress.smac SwitchIngress.smac_hit -m "aa:aa:aa:aa:aa:aa" -a "10"
    ```
    
2. LPM : rib table example。對於LPM, 必須要按照CIDR的形式輸入，否則會出現error.
    ```
     pisc-cli set-flow pipe.SwitchIngress.rib SwitchIngress.hit_route -m "10.0.0.1/24" -a "10"
    ```

3. Ternary : acl example。對於Ternary，value與mask的長度以及形式必須要相同。
    ```
    pisc-cli set-flow pipe.SwitchIngress.acl SwitchIngress.output -m "10/10, 0x0800/0xffff, fa:aa:aa:aa:aa:fa/ff:ff:ff:ff:ff:ff, 127.0.0.1/255.255.255.255, 11.1.1.2/255.255.255.255" -a "0"

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
    ```
3. match : delete specific entry by match key.
    the match flag acts like set-flow function.
    the match flag need to specify talbe's match keys to delete entry.
    ```
    pisc-cli del-flow [TABLE_NAME] -m "Match key, ..."
    pisc-cli del-flow pipe.SwitchIngress.rib -m "192.168.1.7/255.255.255.255, 4000"
    ```

## How to show the table and entries information?
```
pisc-cli [command] [TABLE_NAME]
```
1. table : list all table name.
```
pisc-cli table
```
2. dump :  dump all entries of the table, also -a flag will dump all tables.
```
pisc-cli dump [TABLE_NAME]
pisc-cli dump -a
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