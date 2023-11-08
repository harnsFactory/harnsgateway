**Read this in other languages: [English](README.md), [中文](README_zh.md).**
## What's harnsGateway

HarnsGateway is used to connect industrial equipment.  
It can be installed on embedded devices to as edge industrial gateway.   
It can also be run as a service on the edge IoT platform to as soft gateway.

## What's the functions of harnsGateway

* **Collect equipment data from south end**  
  Supported protocol list:

1. ModbusTcp ModbusRtu ModbusRtuOverTcp
2. S71500
3. OpcUA

* **Control equipment from north end input**
  Supported protocol list:

1. ModbusTcp ModbusRtu ModbusRtuOverTcp

* **Edge computing**

## How to Build

1. git clone https://github.com/harnsFactory/harnsgateway.git
2. cd harnsgateway
3. make
4. cd harnsgateway/_output/bin/

## How to Start

1. QuickStart</br> ./gateway --mqtt-broker-urls=127.0.0.1:1883 &
2. Systemd

## How to Use

example **Connect Modbus device**

1. Started modbus simulator(ModSim32) And update some parameters: deviceId = 1,functionCode = 3,And set first address
   value = 188.</br>[stepOne](https://postimg.cc/sBFyrN2M) </br>Then start service on port 502.
2. Create modbus( [api doc](apis/create-modbus-device.yaml) )device in harnsGateway.The device id property userd in MQTT topic.</br> [stepTow.png](https://postimg.cc/svYFZdpy)
3. Get gateway meta information( [api doc](apis/gateway.yaml) ).The gateway id property used in MQTT topic.</br> [stepThree.png](https://postimg.cc/GHYxf9zP)
4. Subscript MQTT topic.</br> [stepFour.png](https://postimg.cc/ppTGRwqq) </br>Topic is 'data/{gatewayId}/v1/{deviceId}'.
5. Delete the Device.

## How to Run Test


