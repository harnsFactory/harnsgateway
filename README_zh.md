**其他语言版本: [English](README.md), [中文](README_zh.md).**

## harnsGateway是什么

HarnsGateway 可以安装在嵌入式设备上作为硬网关采集工业数据, 也可以安装在边缘系统上面作为软网关采集数据.

## harnsGateway主要功能是什么

* **从南端采集数据**  
  支持的协议列表:

1. ModbusTcp ModbusRtu
2. S71500
3. OpcUA

* **获取北端输入反控设备**

* **边缘计算**

## 如何构建

## 如何启动

## 如何使用

例如 **连接ModbusTcp设备**

1. 打开ModbusTcp模拟软件(ModSim32) 并且更新如下参数: deviceId = 1,functionCode =
   3,并且设置第一个寄存器的值等于188.</br>[stepOne](https://postimg.cc/sBFyrN2M) </br>然后通过502端口启动服务.
2. 在harnsGateway上创建一个ModbusTcp设备.</br> [stepTow.png](https://postimg.cc/svYFZdpy)
3. 订阅MQTT topic.</br> [stepThree.png](https://postimg.cc/ppTGRwqq) </br>Topic为 'data/v1/{deviceId}'.
4. 删除设备.

## 如何启动测试用例

