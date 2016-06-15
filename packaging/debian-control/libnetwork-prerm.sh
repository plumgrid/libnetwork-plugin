# Libnetwork postrm script
#!/bin/sh

set -e

/etc/init.d/libnetwork stop

update-rc.d -f  libnetwork remove

exit 0

