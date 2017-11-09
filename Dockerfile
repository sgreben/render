FROM scratch
ADD bin/linux_amd64/render /render
ENTRYPOINT ["/render"]