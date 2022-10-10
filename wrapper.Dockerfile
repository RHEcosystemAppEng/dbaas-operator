# 0.3.0 catalog image
FROM quay.io/osd-addons/dbaas-operator-index@sha256:83649d689a88763cd2c9e22090af4c9d006deed2d2a6f4ed0a9373811734b1b9

# fix for https://issues.redhat.com/browse/MTSRE-612
RUN chmod u+w /root /usr/bin /usr/lib /usr/sbin /usr/lib64 /usr/lib64/pm-utils
