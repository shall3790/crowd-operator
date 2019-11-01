# FROM registry.access.redhat.com/rhel7:latest
FROM registry.access.redhat.com/ubi7/ubi
# ENV RUN_USER                              daemon
# ENV RUN_GROUP                             daemon
# https://confluence.atlassian.com/display/BitbucketServer/Bitbucket+Server+home+directory
# https://arc-mcac-mlvrh36.amc.faa.gov/artifactory/archive-local/atlassian/crowd/atlassian-crowd-3.7.0.tar.gz
ARG INSTALL_SRC="https://arc-mcac-mlvrh36.amc.faa.gov/artifactory/archive-local"
ARG INSTALL_DIR=/opt/install
ARG INSTALL_FILE_1="atlassian/crowd/atlassian-crowd-3.7.0.tar.gz"
ENV CROWD_VERSION="3.7.0"
ENV CROWD_HOME=/var/crowd-home
ENV CROWD_INSTALL_DIR=/opt/crowd/atlassian-crowd-${CROWD_VERSION}
ENV GIT_SCL=rh-git218
ENV SCL_PKGS="$GIT_SCL rh-python36"
ENV JAVA_HOME=/usr/lib/jvm/java-1.8.0-openjdk-1.8.0.232.b09-0.el7_7.x86_64
ENV GIT_PATH="/opt/rh/$GIT_SCL/root/usr/bin/git"
WORKDIR $CROWD_HOME
# Expose HTTP and SSH ports
EXPOSE 8095
USER 0
# yum repolist --disablerepo=* && \
# subscription-manager repos --enable rhel-server-rhscl-7-rpms && \
# yum --disablerepo=* --enablerepo=rhel-7-server-rpms,rhel-server-rhscl-7-rpms
# RUN yum --disablerepo=* --enablerepo=rhel-7-server-rpms,rhel-server-rhscl-7-rpms install -y java-1.8.0-openjdk java-1.8.0-openjdk-devel nss_wrapper && \
RUN yum install -y java-1.8.0-openjdk java-1.8.0-openjdk-devel nss_wrapper && \
    yum clean all && \
    rm -fr /var/cache/yum
COPY entrypoint.sh                          /entrypoint.sh
COPY crowd-init.properties /tmp/
COPY atlassian-crowd-3.7.0.tar.gz /tmp/
# ARG DOWNLOAD_URL=${INSTALL_SRC}/atlassian/crowd/atlassian-crowd-${CROWD_VERSION}.tar.gz

RUN mkdir -p                                ${CROWD_INSTALL_DIR} ${CROWD_HOME} && \
    tar -xzf /tmp/atlassian-crowd-3.7.0.tar.gz --strip-components=1 -C "${CROWD_INSTALL_DIR}" && \
    cp /tmp/crowd-init.properties ${CROWD_INSTALL_DIR}/crowd-webapp/WEB-INF/classes/ && \
    chown -R 1001:0                      ${CROWD_INSTALL_DIR} ${CROWD_HOME} && \
    chmod -R g=u                         ${CROWD_INSTALL_DIR} ${CROWD_HOME} && \
    chown -R 1001:0                      /entrypoint.sh && \
    chmod -R g=u                         /entrypoint.sh
USER 1001
CMD ["bash", "-c", "$CROWD_INSTALL_DIR/start_crowd.sh -fg"]
ENTRYPOINT ["/entrypoint.sh"]