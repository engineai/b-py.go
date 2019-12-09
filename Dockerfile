FROM python:3.6-alpine

WORKDIR /root
ADD ./pyscripts /root
ENV PYTHONPATH=/root/pyscripts
ENV GOPATH=/go

RUN apk add g++ go \
  && pip install numpy==1.13.3 \
  && rm -rf /var/cache/apk/* /root/.cache /usr/local/lib/python3.6/config-3.6m-x86_64-linux-gnu /usr/local/lib/python3.6/site-packages/pip \
  && python pyscripts/pyscripts.py
