**其他语言版本: [English](README.md), [中文](README_zh.md).**

## harnsGateway是什么

HarnsGateway 可以安装在嵌入式设备上作为硬网关采集工业数据, 也可以安装在边缘系统上面作为软网关采集数据.

## harnsGateway主要功能是什么

* **从南端采集数据**  
  支持的协议列表:

1. ModbusTcp ModbusRtu ModbusRtuOverTcp
2. S71500
3. OpcUA

* **获取北端输入反控设备**
  支持的协议列表:

1. ModbusTcp ModbusRtu ModbusRtuOverTcp
2. S71500

* **边缘计算**

## 如何构建

1. git clone https://github.com/harnsFactory/harnsgateway.git
2. cd harnsgateway
3. make
4. cd harnsgateway/_output/bin/

## 如何启动

1. QuickStart</br> ./gateway --mqtt-broker-urls=127.0.0.1:1883 &
2. Systemd

## 如何使用

例如 **连接Modbus设备**

1. 打开Modbus模拟软件(ModSim32) 并且更新如下参数: deviceId = 1,functionCode =
   3,并且设置第一个寄存器的值等于188.</br>[第一步](https://postimg.cc/sBFyrN2M) </br>然后通过502端口启动服务.
2. 在harnsGateway上创建一个Modbus设备([接口文档](apis/create-modbus-device.yaml)).设备的id属性会在MQTT topic中使用.</br> [第二步](https://postimg.cc/svYFZdpy)
3. 获取harnsGateway元信息( [接口文档](apis/gateway.yaml) ).网关的id属性会在MQTT topic中使用.</br> [第三步](https://postimg.cc/GHYxf9zP)
4. 订阅MQTT topic.</br> [第四步](https://postimg.cc/ppTGRwqq) </br>Topic为'data/{gatewayId}/v1/{deviceId}'.
5. 删除设备.

## 如何启动测试用例


