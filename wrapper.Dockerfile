# 0.2.0 catalog image
FROM quay.io/osd-addons/dbaas-operator-index@sha256:f7bd64974ab6c2d4e055fdd5de7939961db2306e28282350d34c9ddb85bbb50c

# fix for https://issues.redhat.com/browse/MTSRE-612
RUN chmod u+w /root /usr/bin /usr/lib /usr/lib64 /usr/lib64/pm-utils