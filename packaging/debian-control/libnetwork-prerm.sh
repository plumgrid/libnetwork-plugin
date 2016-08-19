# Libnetwork postrm script
#!/bin/sh

set -e

/etc/init.d/libnetwork stop

update-rc.d -f  libnetwork remove
rm -rf /opt/pg/libnetwork/config.ini
rm -rf /opt/pg/libnetwork/libnetwork
exit 0
