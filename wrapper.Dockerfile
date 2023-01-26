# 0.3.0 catalog image
FROM quay.io/osd-addons/dbaas-operator-index@sha256:13a0ef3482b39fe866f4f145faa56b716921a568e27c555590ea315cfb2c18c7

# fix for https://issues.redhat.com/browse/MTSRE-612
RUN chmod u+w /root /usr/bin /usr/lib /usr/sbin /usr/lib64 /usr/lib64/pm-utils
