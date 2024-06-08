# I expected this to list the wif that I just created.
# I think this should be list-wif-configs... but that's a nitpick!
# ...idk think about it, your call:
# ./ocm gcp list-wif-configs
./ocm gcp list-wif-config

# I expected this to provide information about wif with ID 0001
./ocm gcp describe-wif-config 0001

# The help message said ID, but the result was actually derived from the display_name
./ocm gcp describe-wif-config test01
# This wif was listed, but not found?
./ocm gcp describe-wif-config test02

# Will these issues be resolved after the code is connected to the backend instead of mock data?

