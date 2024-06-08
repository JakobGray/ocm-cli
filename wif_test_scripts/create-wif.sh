  ./ocm --debug \
      --log_dir /tmp/test-wif \
      gcp create-wif-config \
      --name test-wif \
      --output-dir /tmp/test-wif \
      --project sda-ccs-1

# Did create-wif actually create the wif resources? 
# It did create it, but linked are ten service accounts instead of the intended
# seven. The additional accounts do not have the "z-" prefix, perhaps this is
# missed in the code?
#
# Where does the public key come from?
