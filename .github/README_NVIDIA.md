# _NVIDIA_


The environment variable `WEBKIT_DISABLE_DMABUF_RENDERER` is used to disable the DMABuf renderer in WebKitGTK, which can help resolve rendering issues on systems with _NVIDIA_ graphics drivers. 

Setting `WEBKIT_DISABLE_DMABUF_RENDERER=1` can be done by exporting it in the terminal before launching an application that uses WebKitGTK. For example, running export WEBKIT_DISABLE_DMABUF_RENDERER=1 before starting an application can bypass the DMABuf renderer and use an alternative rendering method, which may resolve the issue.

This workaround has been reported to work in various scenarios, including issues with Gnome Web, Liferea, and other applications that rely on WebKitGTK. In some cases, additional environment variables such as `WEBKIT_DMABUF_RENDERER_DISABLE_GBM=1` may also be used to further disable specific components of the DMABuf renderer.

It is worth noting that some distributions and packages have implemented downstream patches to automatically disable the DMABuf renderer on _NVIDIA_ hardware, but in certain cases, manually setting the WEBKIT_DISABLE_DMABUF_RENDERER environment variable is still necessary

