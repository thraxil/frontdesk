FROM centurylink/ca-certs
MAINTAINER Anders Pearson <anders@columbia.edu>
COPY frontdesk /
ENV FRONTDESK_DB_PATH /data.db
ENV FRONTDESK_BLEVE_PATH /index.bleve
ENV FRONTDESK_PORT 8000
ENV FRONTDESK_HANDLE_FILE /handles.txt
ENV FRONTDESK_HTPASSWD /htpasswd
EXPOSE 8000
CMD ["/frontdesk"]
