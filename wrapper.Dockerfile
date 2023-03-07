# 0.4.0 catalog image
FROM quay.io/osd-addons/dbaas-operator-index@sha256:2788a47fd0ef1ece30898c1e608050ea71036d3329b9772dbb3d1f69313f745c

# fix for https://issues.redhat.com/browse/MTSRE-612
RUN chmod u+w /root /usr/bin /usr/lib /usr/sbin /usr/lib64 /usr/lib64/pm-utils
