FROM docker.io/bitnami/minideb:bullseye
RUN install_packages curl && \
  mkdir -p /app/bin

WORKDIR /app/bin
ENV PATH="/app/bin:${PATH}"
COPY bin/donkey /app/bin/
CMD ["tail", "-f", "/dev/null"]
