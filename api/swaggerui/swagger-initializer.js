window.onload = function() {
  window.ui = SwaggerUIBundle({
    url: "../_openapi",
    dom_id: '#swagger-ui',
    layout: "BaseLayout"
  });
};
