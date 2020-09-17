#!/usr/bin/env python
# from https://gist.github.com/Thingee/9b3605bfcea928cb0183
# other: https://blog.csdn.net/scaleqiao/article/details/46744891?locationNum=5&fps=1
from rtslib import *

# Setup an IBLOCK backstore
backstore = IBlockBackstore(0, mode='create')

try:
        storage_object = IBlockStorageObject(backstore, "sdb", "/dev/sdb", gen_wwn=True)
except:
        backstore.delete()
            raise

        # Create an iSCSI target endpoint using an iSCSI IQN
        fabric = FabricModule('iscsi')
        target = Target(fabric, "iqn.2003-01.org.linux-iscsi.x.x8664:sn.d3d8b0500fde")
        tpg = TPG(target, 1)

        # Setup a network portal
        portal = NetworkPortal(tpg, "192.168.1.128", "3260")

        # Export LUN
        lun0 = tpg.lun(0, storage_object, "lun")
