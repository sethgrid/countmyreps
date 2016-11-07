FROM centos:7
LABEL authors="Seth Ammons, Jose Lopez"

# build args
#
# file ownership
ARG SYSTEM_USERNAME=countmyreps
ARG SYSTEM_USER_ID=1000
ARG SYSTEM_GROUP_ID=1000
# software versions
ARG GOLANG_VERSION=1.7

# go environment variables
ENV GOPATH /opt/go/
ENV PATH $PATH:/usr/local/go/bin:/opt/go/bin

# install dependencies
RUN yum -y install git make mysql && \
    curl -L --progress https://storage.googleapis.com/golang/go${GOLANG_VERSION}.linux-amd64.tar.gz | tar xz -C /usr/local

# create user
RUN useradd ${SYSTEM_USERNAME} && \
    usermod -u ${SYSTEM_USER_ID} ${SYSTEM_USERNAME} && \
    groupmod -g ${SYSTEM_GROUP_ID} ${SYSTEM_USERNAME}

# copy files over and tweak sample.conf
COPY . /opt/go/src/github.com/sethgrid/countmyreps/.
RUN chown -R ${SYSTEM_USERNAME}:${SYSTEM_USERNAME} /opt/go && \
    sed -i "s~MYSQL_HOST=127.0.0.1~MYSQL_HOST=mysql~g" /opt/go/src/github.com/sethgrid/countmyreps/sample.conf

USER ${SYSTEM_USERNAME}
WORKDIR /opt/go/src/github.com/sethgrid/countmyreps

ENV MYSQL_HOST mysql

RUN go get && go build

CMD /bin/bash -c "source /opt/go/src/github.com/sethgrid/countmyreps/sample.conf && countmyreps"