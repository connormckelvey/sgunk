# Project Dir Structure / Skaffold

- project dir
    - project.toml/json/etc
    - site/
        - index.md
        - other.md
        - blog
            - index.md (optional, can be generated)
            - post.some-post.md

### content file naming pattern:
<type>.[<extra ...>.][<slug> (default = type.String()).]<ext (~ kind)>

eg: index.md ~ index.index.md
    post.1234567890.welcome-to-my-blog.md
    about

# Building the site

### rules
- each file in site directory is a "Page" (Page should be an interface type and various kinds implement certain types of Page)
- every "Page" has or produces content (allowing for virtual/generated pages)
- each page has a type, kind, slug parsed from the source files name
- Directory layout of the built site is 1:1 with the site dir, with the addition of any generated directories
- Processing/building pages should be very similar to a request/response lifecycle of a middleware driven http server
    - Where project, site, and page configuration define a middleware chain for processing a page, the generator builds a context and invokes the middleware chain with the context and a response Writer which completely processes the "Page" into it's final form.

# Page Types
- Page, generic page
- Index, physical or virtual page for a directory index
- Partial, a page partial inteded to be included in another page
- Post, custom type provided by a plugin (like the blog plugin)
- Sitemap, custom virtual page from the sitemap plugin

