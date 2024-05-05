# pdf-go
Golang PDF template library for gofpdf (https://github.com/jung-kurt/gofpdf)

> # NOTE: THIS PROJECT IS A WIP. USE DISCRETION

### What
---

This library adds a templating layer on top of https://github.com/jung-kurt/gofpdf, to ease the pain of having to dealing with the low-level API.
The templates are in JSON format.

NOTE: This code was not created from scratch/ground up. It was rather extracted from an existing system. There may be future improvements and a redesign.

### Why
---

Generating PDFs is challenging. Most services will rely on either web browser driver-based solutions by using HTML for templating, which can be 
unnecesarily costly if all you want is to generate simple uncomplicated PDFs.

So far, https://github.com/jung-kurt/gofpdf (and similar forks) seems to be the only library that supports PDF generation natively without 
external dependencies. It provides a very low-level API which can be cumbersone to work with and this library's aim is to eases the pain of dealing with the low-level API. 

### When
---

Use pdf-go when you want to generate PDF files within your applications, without hvaing to offload to other external systems (which is also still possible).

### How
---

**Anatomy of a template**

All objects inherit from the Element struct, which contains common styling and behavior attributes, some of which are inheritable by default.

At the root is the Document object, then Page, and finally Cell. Everything is rendered inside cells, with individual styling support, hence the template uses
cell-based rendering.


**Reference**

Check [ref.md](ref.md) for reference. 

### To Do

- [ ] Add image support
- [x] Add custom font support
- [x] Alpha channel support (32bit colors)
- [ ] Finish documentation