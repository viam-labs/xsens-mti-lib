# Xsens MTi Library
Library and utilities for connecting to Xsens MTi series IMUs.


Steps to compile in canon:
```sh
sudo apt install -y swig
make clean sdk swig build
```
(you can just run `make build` after the first time)

Steps to run:
```sh
./bin/xsens-mti-lib
```

#Configuration 
Make sure your serial number in your config attributes matches the serial number on the IMU
Configure a local module on your robot with the path to the run.sh in the modules section of the configuration builder.
```
{
"components":
    {
    "model" "viam:sensor:mti-xsens-200",
    "name": "nameItSomething",
    "namespace": "rdk",
    "type": "movement_sensor"
    "attributes" : {
      "serial_path*": "/dev/somethingorother",
      "serial_baud_rate": int, // optional
      "serial_number": "string" // important, check the serial number on the PHYSICAL device and input it here.
      }
    }
  ],
  "modules": [
    {
      "name": "xsens",
      "executable_path": "/path/to/run.sh",
      "type": "local"
    }
}
```
