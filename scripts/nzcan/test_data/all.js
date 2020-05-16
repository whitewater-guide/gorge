
<div id="ALLMap" class="zoomMap embeddedMap"></div>
<script>
    options = {
      InitMap : 'ALLMap',
      BaseHref: LinkTo,
      initPos : {lat: -43.5246182, lng: 172.1109338 }, // {lat: -43.754227, lng: 171.163724 }
      initZoom: 9, 
      minZoom : 6,
      //icon    : 'RiverIcon',
      iconF   : 'riverflow',
      ColourType : 'riverflowColourRange',
      bounds  : {
        north:-41.863194,
        south:-45.078209,
        east: 174.284905,
        west: 169.414539,
      }
    };
  var markers = [{"lat":-42.45731,"lng":172.906357,"SiteName":"Waiau Toa~Clarence River at Jollies (NIWA)","SiteNo":"62105","Value":"7.592 m3\/s","Colour":0,"Total":"0.215","TotalColour":0,"Type":"W"},{"lat":-42.1106262,"lng":173.841934,"SiteName":"Waiau Toa~Clarence River at Clarence Valley Road Bridge","SiteNo":"62107","Value":"0.909 m","Colour":0,"Total":"0.909","TotalColour":0,"Type":"S"},{"lat":-42.368927,"lng":173.67984,"SiteName":"Ashburton SH1","SiteNo":"68801","Value":"0.201 m3\/s","Colour":0,"Total":"0.413","TotalColour":0,"Type":"W"}];
  initMap(markers, options);
</script>