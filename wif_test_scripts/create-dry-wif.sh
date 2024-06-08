# Author: Renan Campos
# Date: 06/07/024 20:01
# Notes:
#  - I expected this local build to create the files for resources of WIF.
#  - I also expected a log file in the log_dir.
  ./ocm --debug \
      --log_dir /tmp/test-wif \
      gcp create-wif-config \
      --dry-run       \
      --name test-wif \
      --output-dir /tmp/test-wif \
      --project sda-ccs-1
