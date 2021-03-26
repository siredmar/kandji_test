package auth

const Redirect = `
<!doctype html>
<html>
<script>
  var hash = window.location.hash;
  var idx = hash.indexOf('#id_token');
  if (idx >= 0) {
    // has '#id_token' in URL
    var at_str = hash.substring(idx);
    var id_token = at_str.substring(at_str.indexOf('=') + 1).trim();
    if (id_token !== "") {
      // we're good to go
      window.location.href = "http://localhost:4445/callback?id_token=" + id_token;
    } else {
      console.error('This page found an empty id_token in the location hash.');
    }
  } else {
    var idx = hash.indexOf('#error=');
    if (idx >= 0) {
      var err_str = hash.substr(idx);
      var end = err_str.indexOf('&');
      if (end < 0) {
        end = err_str.length;
      }
      var err = err_str.substring(err_str.indexOf('=') + 1, end);
      document.write(err);
    } else {
      console.error('This page expected an id_token in the location hash.');
    }
  }
</script>
</html>
`

const Finish = `
<!doctype html>
<html>
<body>
  <img style="width:200px" src="https://uploads-ssl.webflow.com/5f16b7ec2444ee3ebc014407/5f16b87b37350eb2b82bd669_gridX_newlogo_petrol-p-500.png" alt="logo /">
  <h3 style="margin-left: 40px; font-family: Arial, Helvetica, sans-serif">All done!<br /> You can now close this tab.</h3>
</body>
</html>
`
