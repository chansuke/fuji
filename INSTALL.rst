#######################################
MQTT gateway 'Fuji' Install Instruction
#######################################

:version: 0.3.0

Install
============

We have prepared packages to install gateway binary and config file sample easily. You can choose these packages at your environment.


x86_64 Linux
----------------------

#. Download tar.gz file
#. Extract tar.gz file
#. A executable binary will be placed at ./fuji/fuji-gw. A sample config files is ./fuji/config.toml.example.

::

    $ tar xvfz fuji-gw_0.3.0_linux_amd64.tar.gz


For Raspberry Pi (Raspbian)
--------------------------------------

:Package: fuji-gw_0.3.0_raspi_arm6.deb

#. Download a package
#. Install by `dpkg -i`
#. A executable binary will be placed at /usr/local/bin/fuji-gw. A sample config file is /etc/fuji-gw/config.toml.
#. Auto start file is setup as /etc/init.d/fuji-gw.
#. Once you setup /etc/fuji-gw/config.toml correctly, fuji-gw process automatically starts after reboot.

Note: The fuji-gw process runs as a root. 

::

    $ dpkg -i fuji-gw_0.3.0_raspi_arm6.deb
    $ dpkg -i fuji-gw_<Version>_raspi_arm6.deb

For Raspberry Pi 2(Raspbian)
-----------------------------

:Package: fuji-gw_0.3.0_raspi2_arm7.deb

#. Download a package
#. Install by `dpkg -i`
#. A executable binary will be placed at /usr/local/bin/fuji-gw. A sample config file is /etc/fuji-gw/config.toml.

::

    $ dpkg -i fuji-gw_0.3.0_raspi2_arm7.deb
    $ fuji-gw -c /etc/fuji-gw/config.toml

For Armadillo-IoT
--------------------

:tar.gz File: fuji-gw_0.3.0_arm5.tar.gz

(Note:) If you want to embed Fuji to startup image, please refer `Armadillo-IoT Gateway Standard manual <http://manual.atmark-techno.com/armadillo-iot/armadillo-iotg-std_product_manual_ja-1.1.1/>`_ of Atmark-Techno.

#. Download tar.gz on your PC
#. Extract tar.gz to the current directory on your PC
#. Connect Armadilli-IoT and your PC via serial cable, run your terminal software, then start up armadillo-IoT and login. Please read more detail from `Manual <http://manual.atmark-techno.com/armadillo-iot/armadillo-iotg-std_product_manual_ja-1.1.1/ch05.html#sec-login>`_
#. After Armadillo-IoT started up, check IPaddress settings on the Armadillio-IoT by using ifconfig command.
#. Send a executable binary and a config file from your PC via ftp. Please use IPaddress of you just confirmed. The binary and the config file will be sent to `/home/ftp/pub`.
#. Execute fuji-gw from serial console.

::

    $ wget <tar.gz URL>
    $ tar zxf fuji-gw_<Version>_arm5.tar.gz
    $ cd fuji

    ftp <Armadillo-IoT IPAddress>
    ftp> Name: ftp
    ftp> Password: (password is empty)
    ftp> cd pub
    ftp> binary
    ftp> put fuji-gw
    ftp> put config.toml.example

For Intel Edison
-------------------

:Package: fuji-gw_0.3.0_edison_386.ipk

#. Login to Intel Edison
#. Download package          
#. Install using `opkg install` command.
#. A executable binary will be placed at ./fuji/fuji-gw. A sample config files is ./fuji/config.toml.example.

::

    $ wget <packge url>
    $ opkg install fuji-gw_<Version>_edison_i386.ipk

Build from source code
------------------------------

#. Prepare go development environment.
#. Do `go get` to install fuji command.

::

   $ go get github.com/shiguredo/fuji/cmd/fuji


Config example
==============

A MQTT broker is required to check the Fuji is running. At this example, we will use `Sango <https://sango.shiguredo.jp>`_ which is an `MQTT as a Servce` produced by Shiguredo.

Before setting up Fuji, please access the Sango page and Sign up using your own GitHub account.


Config file
------------------

In this config file, we use dummy device function of Fuji. A dummy device can send some static data to MQTT Broker same as an Sensor.
Since all user can use only topic under `<username>/#` on the Sango, set `topic_prefix` value as is.

.. code-block:: toml

    [gateway]
    
        name = "fuji"
    
    [[broker."sango"]]
    
        host = "<sango hostname>"
        port = 1883
    
        username = "<sango username>"
        password = "<sango password>"
    
        retry_interval = 10
        topic_prefix = "<sango username>"
    
    
    [device."test/dummy"]
    
        broker = "sango"
        qos = 0
    
        interval = 10
        payload = "Hello MQTT."
    
        type = Dummy

Then, execute fuji-gw with the config file.

::

    $ ./fuji-gw -c <config file path>


Config example
^^^^^^^^^^^^^^^^^^

This example is set like below.

- host: sango.example.com
- username: shiguredo
- password: pass


.. code-block:: toml

    [gateway]
    
        name = "fuji"
    
    [[broker."sango"]]
    
        host = "sango.example.com"
        port = 1883
    
        username = "shiguredo"
        password = "pass"
    
        retry_interval = 10
        topic_prefix = "shiguredo@github"
    
    
    [device."test/dummy"]
    
        broker = "sango"
        qos = 0
    
        interval = 10
        payload = "Hello MQTT."
    
        type = "Dummy"
    


Operation check using mqttcli
------------------------------------------------

`mqttcli` is an tool which can Publish or Subscribe from command line easily.
You can download from `this page <https://drone.io/github.com/shirou/mqttcli/files>`_ . There are some binary which includes Windows, Mac OS, Intel Edison and so on.

After download mqttcli, create setting file for mqttcli.

settings.json::

    {
      "host": "<sango hostname>",
      "port": 1883,
      "username": "<sango username>",
      "password": "<sango password>"
    }


command::

    $ mqttcli sub --conf settings.json  -t "<sango username>/<fuji gateway name>/<device name>/<device type>"


example::

    $ mqttcli sub --conf settings.json  -t "shiguredo@github/fuji/test/Dummy"

If you confirm "Hello MQTT." message is sent every 10 sec, that's it.

