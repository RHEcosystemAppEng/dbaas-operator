# 0.2.0 catalog image
FROM quay.io/osd-addons/dbaas-operator-index@sha256:b699851c2a839ee85a98a8daf3b619c0b34716c081046b229c37e8ea2d2efa96

# fix for https://issues.redhat.com/browse/MTSRE-612
RUN chmod u+w /root /usr/bin /usr/lib /usr/lib64 /usr/lib64/pm-utils