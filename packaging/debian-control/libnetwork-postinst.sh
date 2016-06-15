# Libnetwork postinst script
#!/bin/sh

set -e

update-rc.d libnetwork defaults
update-rc.d libnetwork enable

/etc/init.d/libnetwork start

exit 0

