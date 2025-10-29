FROM scratch
COPY tbonsai /usr/bin/tbonsai
ENV HOME=/home/user
ENTRYPOINT ["/usr/bin/tbonsai"]
