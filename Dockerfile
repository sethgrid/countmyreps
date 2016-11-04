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

ENV PATH $PATH:/usr/local/go/bin
# export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
# ENV GOPATH /opt/countmyreps

# install dependencies
RUN yum -y install git make
RUN curl -L --progress https://storage.googleapis.com/golang/go${GOLANG_VERSION}.linux-amd64.tar.gz | tar xz -C /usr/local

# create user
RUN useradd ${SYSTEM_USERNAME} && \
    usermod -u ${SYSTEM_USER_ID} ${SYSTEM_USERNAME} && \
    groupmod -g ${SYSTEM_GROUP_ID} ${SYSTEM_USERNAME}

RUN mkdir -p /opt/go/countmyreps && \
    chown -R ${SYSTEM_USERNAME}:${SYSTEM_USERNAME} /opt/go

COPY . /opt/go/countmyreps/.

USER ${SYSTEM_USERNAME}
ENV GOPATH $HOME
WORKDIR /opt/go/countmyreps

RUN go get
RUN go build

ENTRYPOINT ./build/countmyreps