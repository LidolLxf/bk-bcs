FROM tencentos/tencentos4-minimal

#for command envsubst
RUN yum install -y gettext

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo 'LANG="en_US.UTF-8"' > /etc/locale.conf
ENV LANG=en_US.UTF-8 \
    LANGUAGE=en_US.UTF-8

RUN mkdir -p /data/bcs/bcs-powertrading/
RUN mkdir -p /data/bcs/logs/bcs

ADD bcs-powertrading /data/bcs/bcs-powertrading/
ADD bcs-powertrading.json.template /data/bcs/bcs-powertrading/
ADD container-start.sh /data/bcs/bcs-powertrading/

RUN chmod +x /data/bcs/bcs-powertrading/container-start.sh
RUN chmod +x /data/bcs/bcs-powertrading/bcs-powertrading


WORKDIR /data/bcs/bcs-powertrading/
CMD [ "/data/bcs/bcs-powertrading/container-start.sh" ]
