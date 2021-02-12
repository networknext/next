
Hosting in AWS
==============

Network Next can now autodetect the datacenter when you are hosting in AWS.

To use this autodetection, simply pass in a datacenter of "" to _next_server_create_ when you run on AWS, while making sure that NEXT_DATACENTER environment variable is also not set, as it overrides whatever datacenter string you pass in.

Important note. On AWS, it's not possible to look up the network next datacenter directly from the availability zone, because it's potentially different on different customer accounts. Instead, we use the account independent AZID, which is guaranteed to be the same across all accounts.

Please see this article for more details: https://docs.aws.amazon.com/ram/latest/userguide/working-with-az-ids.html

The mapping from AZID to datacenter string is as follows:

========================== =========================================
AZID                       Network Next Datacenter
-------------------------- -----------------------------------------
use1-az1                   amazon.virginia.1
use1-az2                   amazon.virginia.2
use1-az3                   amazon.virginia.3
use1-az4                   amazon.virginia.4
use1-az5                   amazon.virginia.5
use1-az6                   amazon.virginia.6
use2-az1                   amazon.ohio.1
use2-az2                   amazon.ohio.2
use2-az3                   amazon.ohio.3
usw1-az1                   amazon.sanjose.1
usw1-az3                   amazon.sanjose.3
usw2-az1                   amazon.oregon.1
usw2-az2                   amazon.oregon.2
usw2-az3                   amazon.oregon.3
usw2-az4                   amazon.oregon.4
ape1-az1                   amazon.hongkong.1
ape1-az2                   amazon.hongkong.2
ape1-az3                   amazon.hongkong.3
aps1-az1                   amazon.mumbai.1
aps1-az2                   amazon.mumbai.2
aps1-az3                   amazon.mumbai.3
apne2-az1                  amazon.seoul.1
apne2-az2                  amazon.seoul.2
apne2-az3                  amazon.seoul.3
apne2-az4                  amazon.seoul.4
apse1-az1                  amazon.singapore.1
apse1-az2                  amazon.singapore.2
apse1-az3                  amazon.singapore.3
apse2-az1                  amazon.sydney.1
apse2-az2                  amazon.sydney.2
apse2-az3                  amazon.sydney.3
apne1-az1                  amazon.tokyo.1
apne1-az2                  amazon.tokyo.2
apne1-az4                  amazon.tokyo.4
cac1-az4                   amazon.montreal.1
cac1-az2                   amazon.montreal.2
cac1-az3                   amazon.montreal.4
euc1-az1                   amazon.frankfurt.1
euc1-az2                   amazon.frankfurt.2
euc1-az3                   amazon.frankfurt.3
euw1-az1                   amazon.dublin.1
euw1-az2                   amazon.dublin.2
euw1-az3                   amazon.dublin.3
euw2-az1                   amazon.london.1
euw2-az2                   amazon.london.2
euw2-az3                   amazon.london.3
eus1-az1                   amazon.milan.1
eus1-az2                   amazon.milan.2
eus1-az3                   amazon.milan.3
euw3-az1                   amazon.paris.1
euw3-az2                   amazon.paris.2
euw3-az3                   amazon.paris.3
eun1-az1                   amazon.stockholm.1
eun1-az2                   amazon.stockholm.2
eun1-az3                   amazon.stockholm.3
mes1-az1                   amazon.bahrain.1
mes1-az2                   amazon.bahrain.2
mes1-az3                   amazon.bahrain.3
sae1-az1                   amazon.saopaulo.1
sae1-az2                   amazon.saopaulo.2
sae1-az3                   amazon.saopaulo.3
afs1-az1                   amazon.capetown.1
afs1-az2                   amazon.capetown.2
afs1-az3                   amazon.capetown.3
========================== =========================================

https://storage.googleapis.com/network-next-sdk/amazon.txt