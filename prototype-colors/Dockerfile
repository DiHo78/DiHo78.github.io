# PROTOTYPE, throwaway: Hugo extended + Go + Dart Sass for the converter comparison.
FROM golang:1.25-bookworm
ARG HUGO_VERSION=0.169.0
ARG HUGO_SHA256=c41672a0725d307758a132bc94a962e59760eb33f9ba08c991b81f9881d6fa7d
# Stack calls toCSS, which means libsass, deprecated since Hugo 0.153: the warning breaks --panicOnWarning.
# site/layouts/_partials/head/style.html switches to Dart Sass, which needs this binary.
ARG SASS_VERSION=1.107.1
ARG SASS_SHA256=6cefe78692d03ca53843394bb98de4dccf9a3d8ee5ece40be7316be852624003
RUN curl -fsSL -o /tmp/hugo.tgz \
      "https://github.com/gohugoio/hugo/releases/download/v${HUGO_VERSION}/hugo_extended_${HUGO_VERSION}_linux-amd64.tar.gz" \
 && echo "${HUGO_SHA256}  /tmp/hugo.tgz" | sha256sum -c - \
 && tar -xzf /tmp/hugo.tgz -C /usr/local/bin hugo \
 && curl -fsSL -o /tmp/sass.tgz \
      "https://github.com/sass/dart-sass/releases/download/${SASS_VERSION}/dart-sass-${SASS_VERSION}-linux-x64.tar.gz" \
 && echo "${SASS_SHA256}  /tmp/sass.tgz" | sha256sum -c - \
 && tar -xzf /tmp/sass.tgz -C /opt \
 && ln -s /opt/dart-sass/sass /usr/local/bin/sass \
 && rm /tmp/hugo.tgz /tmp/sass.tgz \
 && hugo version && sass --version
WORKDIR /site
