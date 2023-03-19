FROM alpine

COPY mino /root/mino

RUN chmod +x /root/mino && mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

RUN echo 'address: ":28648"' >> /root/mino.yml
RUN echo 'log_enable: false' >> /root/mino.yml

WORKDIR /root

EXPOSE 28648

CMD ["/root/mino"]