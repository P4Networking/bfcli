

```bash
pisc-cli set-flow pipe.SwitchIngress.rmac SwitchIngress.rmac_hit -m "ff:ff:ff:ff:ff:ff" -a ""
pisc-cli set-flow pipe.SwitchIngress.dmac SwitchIngress.dmac_hit -m "ff:ff:ff:ff:ff:ff" -a "511"
pisc-cli set-flow pipe.SwitchIngress.fib  SwitchIngress.hit_route_port -m "255.255.255.255" -a "511"
pisc-cli set-flow pipe.SwitchIngress.rib_24 SwitchIngress.hit_route_port -m "255.255.255" -a "511"
pisc-cli set-flow pipe.SwitchIngress.rib_20 SwitchIngress.hit_route_port -m "15.255.255" -a "511"
pisc-cli set-flow pipe.SwitchIngress.rib_16 SwitchIngress.hit_route_port -m "255.255" -a "511"
pisc-cli set-flow pipe.SwitchIngress.rib_12 SwitchIngress.hit_route_port -m "15.255" -a "511"
pisc-cli set-flow pipe.SwitchIngress.rib_8 SwitchIngress.hit_route_port -m "255" -a "511"
pisc-cli set-flow pipe.SwitchIngress.rib SwitchIngress.hit_route_port -m "255.255.255.255/255.255.255.255, 4000" -a "511"
pisc-cli set-flow pipe.SwitchIngress.nexthop SwitchIngress.set_nexthop -m "255" -a "511, ff:ff:ff:ff:ff:ff"
pisc-cli set-flow pipe.SwitchEgress.smac_rewrite_by_portid SwitchEgress.rewrite_smac -m "511/511, 4000" -a "ff:ff:ff:ff:ff:ff"
pisc-cli set-flow pipe.SwitchIngress.arp SwitchIngress.arp_response -m "511/511, 0xffff, 0xffff, 255.255.255.255/255.255.255.255, 4000" -a "ff:ff:ff:ff:ff:ff"
pisc-cli set-flow pipe.SwitchIngress.acl SwitchIngress.output -m "511/511, 0xffff/0xffff, aa:aa:aa:aa:aa:aa/ff:ff:ff:ff:ff:ff, 255.255.255.255/255.255.255.255, 255.255.255.255/255.255.255.255, 255/255, 4000" -a "10"
pisc-cli set-flow pipe.SwitchIngress.smac SwitchIngress.smac_hit -m "ff:ff:ff:ff:ff:ff" -a "511" -t 600
```


```bash
pisc-cli del-flow pipe.SwitchIngress.rmac -m "ff:ff:ff:ff:ff:ff"
pisc-cli del-flow pipe.SwitchIngress.dmac -m "ff:ff:ff:ff:ff:ff"
pisc-cli del-flow pipe.SwitchIngress.fib -m "255.255.255.255" 
pisc-cli del-flow pipe.SwitchIngress.rib_24 -m "255.255.255" 
pisc-cli del-flow pipe.SwitchIngress.rib_20 -m "15.255.255"
pisc-cli del-flow pipe.SwitchIngress.rib_16 -m "255.255"
pisc-cli del-flow pipe.SwitchIngress.rib_12 -m "15.255"
pisc-cli del-flow pipe.SwitchIngress.rib_8 -m "255" 
pisc-cli del-flow pipe.SwitchIngress.rib -m "255.255.255.255/255.255.255.255, 4000" 
pisc-cli del-flow pipe.SwitchIngress.nexthop -m "255"
pisc-cli del-flow pipe.SwitchEgress.smac_rewrite_by_portid -m "511/511, 4000" 
pisc-cli del-flow pipe.SwitchIngress.arp -m "511/511, 0xffff, 0xffff, 255.255.255.255/255.255.255.255, 4000"
pisc-cli del-flow pipe.SwitchIngress.acl -m "511/511, 0xffff/0xffff, aa:aa:aa:aa:aa:aa/ff:ff:ff:ff:ff:ff, 255.255.255.255/255.255.255.255, 255.255.255.255/255.255.255.255, 255/255, 4000"
pisc-cli del-flow pipe.SwitchIngress.smac -m "ff:ff:ff:ff:ff:ff"
```