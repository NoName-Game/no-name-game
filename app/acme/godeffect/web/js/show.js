THREE.TrackballControls = function(object, domElement) {
  var _this = this;
  var STATE = {
    NONE: -1,
    ROTATE: 0,
    ZOOM: 1,
    PAN: 2,
    TOUCH_ROTATE: 3,
    TOUCH_ZOOM_PAN: 4
  };

  this.object = object;
  this.domElement = domElement !== undefined ? domElement : document;

  // API

  this.enabled = true;

  this.screen = { left: 0, top: 0, width: 0, height: 0 };

  this.rotateSpeed = 1.0;
  this.zoomSpeed = 1.2;
  this.panSpeed = 1.5;

  this.noRotate = false;
  this.noZoom = false;
  this.noPan = false;

  this.staticMoving = false;
  this.dynamicDampingFactor = 0.2;

  this.minDistance = 0;
  this.maxDistance = Infinity;

  this.keys = [65 /*A*/, 83 /*S*/, 68 /*D*/];

  // internals

  this.target = new THREE.Vector3();

  var EPS = 0.000001;

  var lastPosition = new THREE.Vector3();

  var _state = STATE.NONE,
    _prevState = STATE.NONE,
    _eye = new THREE.Vector3(),
    _movePrev = new THREE.Vector2(),
    _moveCurr = new THREE.Vector2(),
    _lastAxis = new THREE.Vector3(),
    _lastAngle = 0,
    _zoomStart = new THREE.Vector2(),
    _zoomEnd = new THREE.Vector2(),
    _touchZoomDistanceStart = 0,
    _touchZoomDistanceEnd = 0,
    _panStart = new THREE.Vector2(),
    _panEnd = new THREE.Vector2();

  // for reset

  this.target0 = this.target.clone();
  this.position0 = this.object.position.clone();
  this.up0 = this.object.up.clone();

  // events

  var changeEvent = { type: "change" };
  var startEvent = { type: "start" };
  var endEvent = { type: "end" };

  // methods

  this.handleResize = function() {
    if (this.domElement === document) {
      this.screen.left = 0;
      this.screen.top = 0;
      this.screen.width = window.innerWidth;
      this.screen.height = window.innerHeight;
    } else {
      var box = this.domElement.getBoundingClientRect();
      // adjustments come from similar code in the jquery offset() function
      var d = this.domElement.ownerDocument.documentElement;
      this.screen.left = box.left + window.pageXOffset - d.clientLeft;
      this.screen.top = box.top + window.pageYOffset - d.clientTop;
      this.screen.width = box.width;
      this.screen.height = box.height;
    }
  };

  this.handleEvent = function(event) {
    if (typeof this[event.type] == "function") {
      this[event.type](event);
    }
  };

  var getMouseOnScreen = (function() {
    var vector = new THREE.Vector2();
    return function getMouseOnScreen(pageX, pageY) {
      vector.set(
        (pageX - _this.screen.left) / _this.screen.width,
        (pageY - _this.screen.top) / _this.screen.height
      );
      return vector;
    };
  })();

  var getMouseOnCircle = (function() {
    var vector = new THREE.Vector2();
    return function getMouseOnCircle(pageX, pageY) {
      vector.set(
        (pageX - _this.screen.width * 0.5 - _this.screen.left) /
          (_this.screen.width * 0.5),
        (_this.screen.height + 2 * (_this.screen.top - pageY)) /
          _this.screen.width // screen.width intentional
      );
      return vector;
    };
  })();

  this.rotateCamera = (function() {
    var axis = new THREE.Vector3(),
      quaternion = new THREE.Quaternion(),
      eyeDirection = new THREE.Vector3(),
      objectUpDirection = new THREE.Vector3(),
      objectSidewaysDirection = new THREE.Vector3(),
      moveDirection = new THREE.Vector3(),
      angle;

    return function rotateCamera() {
      moveDirection.set(
        _moveCurr.x - _movePrev.x,
        _moveCurr.y - _movePrev.y,
        0
      );
      angle = moveDirection.length();
      if (angle) {
        _eye.copy(_this.object.position).sub(_this.target);
        eyeDirection.copy(_eye).normalize();
        objectUpDirection.copy(_this.object.up).normalize();
        objectSidewaysDirection
          .crossVectors(objectUpDirection, eyeDirection)
          .normalize();
        objectUpDirection.setLength(_moveCurr.y - _movePrev.y);
        objectSidewaysDirection.setLength(_moveCurr.x - _movePrev.x);
        moveDirection.copy(objectUpDirection.add(objectSidewaysDirection));
        axis.crossVectors(moveDirection, _eye).normalize();
        angle *= _this.rotateSpeed;
        quaternion.setFromAxisAngle(axis, angle);
        _eye.applyQuaternion(quaternion);
        _this.object.up.applyQuaternion(quaternion);
        _lastAxis.copy(axis);
        _lastAngle = angle;
      } else if (!_this.staticMoving && _lastAngle) {
        _lastAngle *= Math.sqrt(1.0 - _this.dynamicDampingFactor);
        _eye.copy(_this.object.position).sub(_this.target);
        quaternion.setFromAxisAngle(_lastAxis, _lastAngle);
        _eye.applyQuaternion(quaternion);
        _this.object.up.applyQuaternion(quaternion);
      }
      _movePrev.copy(_moveCurr);
    };
  })();

  this.zoomCamera = function() {
    var factor;
    if (_state === STATE.TOUCH_ZOOM_PAN) {
      factor = _touchZoomDistanceStart / _touchZoomDistanceEnd;
      _touchZoomDistanceStart = _touchZoomDistanceEnd;
      _eye.multiplyScalar(factor);
    } else {
      factor = 1.0 + (_zoomEnd.y - _zoomStart.y) * _this.zoomSpeed;
      if (factor !== 1.0 && factor > 0.0) {
        _eye.multiplyScalar(factor);
        if (_this.staticMoving) {
          _zoomStart.copy(_zoomEnd);
        } else {
          _zoomStart.y +=
            (_zoomEnd.y - _zoomStart.y) * this.dynamicDampingFactor;
        }
      }
    }
  };

  this.panCamera = (function() {
    var mouseChange = new THREE.Vector2(),
      objectUp = new THREE.Vector3(),
      pan = new THREE.Vector3();

    return function panCamera() {
      mouseChange.copy(_panEnd).sub(_panStart);
      if (mouseChange.lengthSq()) {
        mouseChange.multiplyScalar(_eye.length() * _this.panSpeed);
        pan
          .copy(_eye)
          .cross(_this.object.up)
          .setLength(mouseChange.x);
        pan.add(objectUp.copy(_this.object.up).setLength(mouseChange.y));
        _this.object.position.add(pan);
        _this.target.add(pan);
        if (_this.staticMoving) {
          _panStart.copy(_panEnd);
        } else {
          _panStart.add(
            mouseChange
              .subVectors(_panEnd, _panStart)
              .multiplyScalar(_this.dynamicDampingFactor)
          );
        }
      }
    };
  })();

  this.checkDistances = function() {
    if (!_this.noZoom || !_this.noPan) {
      if (_eye.lengthSq() > _this.maxDistance * _this.maxDistance) {
        _this.object.position.addVectors(
          _this.target,
          _eye.setLength(_this.maxDistance)
        );
        _zoomStart.copy(_zoomEnd);
      }
      if (_eye.lengthSq() < _this.minDistance * _this.minDistance) {
        _this.object.position.addVectors(
          _this.target,
          _eye.setLength(_this.minDistance)
        );
        _zoomStart.copy(_zoomEnd);
      }
    }
  };

  this.update = function() {
    _eye.subVectors(_this.object.position, _this.target);
    if (!_this.noRotate) {
      _this.rotateCamera();
    }
    if (!_this.noZoom) {
      _this.zoomCamera();
    }
    if (!_this.noPan) {
      _this.panCamera();
    }
    _this.object.position.addVectors(_this.target, _eye);
    _this.checkDistances();
    _this.object.lookAt(_this.target);
    if (lastPosition.distanceToSquared(_this.object.position) > EPS) {
      _this.dispatchEvent(changeEvent);
      lastPosition.copy(_this.object.position);
    }
  };

  this.reset = function() {
    _state = STATE.NONE;
    _prevState = STATE.NONE;
    _this.target.copy(_this.target0);
    _this.object.position.copy(_this.position0);
    _this.object.up.copy(_this.up0);
    _eye.subVectors(_this.object.position, _this.target);
    _this.object.lookAt(_this.target);
    _this.dispatchEvent(changeEvent);
    lastPosition.copy(_this.object.position);
  };

  // listeners

  function keydown(event) {
    if (_this.enabled === false) return;

    window.removeEventListener("keydown", keydown);
    _prevState = _state;
    if (_state !== STATE.NONE) {
      return;
    } else if (event.keyCode === _this.keys[STATE.ROTATE] && !_this.noRotate) {
      _state = STATE.ROTATE;
    } else if (event.keyCode === _this.keys[STATE.ZOOM] && !_this.noZoom) {
      _state = STATE.ZOOM;
    } else if (event.keyCode === _this.keys[STATE.PAN] && !_this.noPan) {
      _state = STATE.PAN;
    }
  }

  function keyup(event) {
    if (_this.enabled === false) return;
    _state = _prevState;
    window.addEventListener("keydown", keydown, false);
  }

  function mousedown(event) {
    if (_this.enabled === false) return;

    event.preventDefault();
    event.stopPropagation();

    if (_state === STATE.NONE) {
      _state = event.button;
    }
    if (_state === STATE.ROTATE && !_this.noRotate) {
      _moveCurr.copy(getMouseOnCircle(event.pageX, event.pageY));
      _movePrev.copy(_moveCurr);
    } else if (_state === STATE.ZOOM && !_this.noZoom) {
      _zoomStart.copy(getMouseOnScreen(event.pageX, event.pageY));
      _zoomEnd.copy(_zoomStart);
    } else if (_state === STATE.PAN && !_this.noPan) {
      _panStart.copy(getMouseOnScreen(event.pageX, event.pageY));
      _panEnd.copy(_panStart);
    }

    document.addEventListener("mousemove", mousemove, false);
    document.addEventListener("mouseup", mouseup, false);

    _this.dispatchEvent(startEvent);
  }

  function mousemove(event) {
    if (_this.enabled === false) return;

    event.preventDefault();
    event.stopPropagation();

    if (_state === STATE.ROTATE && !_this.noRotate) {
      _movePrev.copy(_moveCurr);
      _moveCurr.copy(getMouseOnCircle(event.pageX, event.pageY));
    } else if (_state === STATE.ZOOM && !_this.noZoom) {
      _zoomEnd.copy(getMouseOnScreen(event.pageX, event.pageY));
    } else if (_state === STATE.PAN && !_this.noPan) {
      _panEnd.copy(getMouseOnScreen(event.pageX, event.pageY));
    }
  }

  function mouseup(event) {
    if (_this.enabled === false) return;

    event.preventDefault();
    event.stopPropagation();

    _state = STATE.NONE;

    document.removeEventListener("mousemove", mousemove);
    document.removeEventListener("mouseup", mouseup);
    _this.dispatchEvent(endEvent);
  }

  function mousewheel(event) {
    if (_this.enabled === false) return;

    event.preventDefault();
    event.stopPropagation();

    var delta = 0;

    if (event.wheelDelta) {
      // WebKit / Opera / Explorer 9
      delta = event.wheelDelta / 40;
    } else if (event.detail) {
      // Firefox
      delta = -event.detail / 3;
    }
    _zoomStart.y += delta * 0.01;
    _this.dispatchEvent(startEvent);
    _this.dispatchEvent(endEvent);
  }

  function touchstart(event) {
    if (_this.enabled === false) return;

    switch (event.touches.length) {
      case 1:
        _state = STATE.TOUCH_ROTATE;
        _moveCurr.copy(
          getMouseOnCircle(event.touches[0].pageX, event.touches[0].pageY)
        );
        _movePrev.copy(_moveCurr);
        break;
      default:
        // 2 or more
        _state = STATE.TOUCH_ZOOM_PAN;
        var dx = event.touches[0].pageX - event.touches[1].pageX;
        var dy = event.touches[0].pageY - event.touches[1].pageY;
        _touchZoomDistanceEnd = _touchZoomDistanceStart = Math.sqrt(
          dx * dx + dy * dy
        );

        var x = (event.touches[0].pageX + event.touches[1].pageX) / 2;
        var y = (event.touches[0].pageY + event.touches[1].pageY) / 2;
        _panStart.copy(getMouseOnScreen(x, y));
        _panEnd.copy(_panStart);
        break;
    }
    _this.dispatchEvent(startEvent);
  }

  function touchmove(event) {
    if (_this.enabled === false) return;

    event.preventDefault();
    event.stopPropagation();

    switch (event.touches.length) {
      case 1:
        _movePrev.copy(_moveCurr);
        _moveCurr.copy(
          getMouseOnCircle(event.touches[0].pageX, event.touches[0].pageY)
        );
        break;
      default:
        // 2 or more
        var dx = event.touches[0].pageX - event.touches[1].pageX;
        var dy = event.touches[0].pageY - event.touches[1].pageY;
        _touchZoomDistanceEnd = Math.sqrt(dx * dx + dy * dy);

        var x = (event.touches[0].pageX + event.touches[1].pageX) / 2;
        var y = (event.touches[0].pageY + event.touches[1].pageY) / 2;
        _panEnd.copy(getMouseOnScreen(x, y));
        break;
    }
  }

  function touchend(event) {
    if (_this.enabled === false) return;

    switch (event.touches.length) {
      case 0:
        _state = STATE.NONE;
        break;
      case 1:
        _state = STATE.TOUCH_ROTATE;
        _moveCurr.copy(
          getMouseOnCircle(event.touches[0].pageX, event.touches[0].pageY)
        );
        _movePrev.copy(_moveCurr);
        break;
    }
    _this.dispatchEvent(endEvent);
  }

  function contextmenu(event) {
    event.preventDefault();
  }

  this.dispose = function() {
    this.domElement.removeEventListener("contextmenu", contextmenu, false);
    this.domElement.removeEventListener("mousedown", mousedown, false);
    this.domElement.removeEventListener("mousewheel", mousewheel, false);
    this.domElement.removeEventListener(
      "MozMousePixelScroll",
      mousewheel,
      false
    ); // firefox

    this.domElement.removeEventListener("touchstart", touchstart, false);
    this.domElement.removeEventListener("touchend", touchend, false);
    this.domElement.removeEventListener("touchmove", touchmove, false);

    document.removeEventListener("mousemove", mousemove, false);
    document.removeEventListener("mouseup", mouseup, false);

    window.removeEventListener("keydown", keydown, false);
    window.removeEventListener("keyup", keyup, false);
  };

  this.domElement.addEventListener("contextmenu", contextmenu, false);
  this.domElement.addEventListener("mousedown", mousedown, false);
  this.domElement.addEventListener("mousewheel", mousewheel, false);
  this.domElement.addEventListener("MozMousePixelScroll", mousewheel, false); // firefox

  this.domElement.addEventListener("touchstart", touchstart, false);
  this.domElement.addEventListener("touchend", touchend, false);
  this.domElement.addEventListener("touchmove", touchmove, false);

  window.addEventListener("keydown", keydown, false);
  window.addEventListener("keyup", keyup, false);

  this.handleResize();

  // force an update at start
  this.update();
};

THREE.TrackballControls.prototype = Object.create(
  THREE.EventDispatcher.prototype
);
THREE.TrackballControls.prototype.constructor = THREE.TrackballControls;

var container, stats;
var camera, scene, renderer, controls;
var GGrid;

var galaxy;
var selectedStar;
var locator;

var raycaster;
var mouse;
var INTERSECTED;

var stars = [];
var starsColor = [];
var starList = [];
var nbStar = 0;

//initial data
var maxStar = 10000;
var galaxySize = 6000;
var galaxyArm = 2;
var galaxyRoll = 360;

// generated Data
var firstSeed = 0;
var nbHS = 0;
var nbHP = 0;
var nbNHP = 0;

var runLoopIdentifier;

$(document).ready(function() {
  setScene();
  animate();
  $(".loader").hide();
  $(".extras").show();

  streamStar();
});

function streamStar(nextSeed = false) {
  var mydata = JSON.parse(
    '[{"position":{"x":22.16307806080818,"y":-11.977081910933558,"z":-1.2169770157727515},"size":null,"temperature":5380.147834951127},{"position":{"x":-16.142240455015916,"y":2.639755853212292,"z":-20.103831128020868},"size":null,"temperature":3714.120908041541},{"position":{"x":19.15559691797205,"y":13.810310023346124,"z":-5.1361995936588185},"size":null,"temperature":5308.019913876438},{"position":{"x":-135.81275241422125,"y":3.0005843404214865,"z":-113.58279320243943},"size":null,"temperature":6333.650395895192},{"position":{"x":185.48610355407115,"y":169.88136125163726,"z":-32.563649277671686},"size":null,"temperature":5791.4940737148245},{"position":{"x":-43.249015778759144,"y":-132.13923800167268,"z":-70.64633092962346},"size":null,"temperature":1060.6452892816835},{"position":{"x":62.56219668131926,"y":22.58432412092343,"z":61.429997805773255},"size":null,"temperature":2943.0767553593387},{"position":{"x":-21.255372183275945,"y":-159.45878491382555,"z":-22.981929379372982},"size":null,"temperature":1028.4153348898587},{"position":{"x":6.672363120488466,"y":-21.879352295668458,"z":27.031684085584818},"size":null,"temperature":2936.465083591857},{"position":{"x":-76.40836735877389,"y":16.741121888215137,"z":-171.05950720384033},"size":null,"temperature":1437.4111315409705},{"position":{"x":136.30169679542834,"y":-57.318260701591655,"z":-82.51296556226713},"size":null,"temperature":3196.055200973551},{"position":{"x":-20.6813240721745,"y":149.65708638237402,"z":-53.08650936658856},"size":null,"temperature":3546.520677649658},{"position":{"x":1.3883110520216668,"y":-14.870610991367496,"z":31.710655509622427},"size":null,"temperature":4982.456131364432},{"position":{"x":-27.695163821228064,"y":-69.26264848038188,"z":40.01601319341083},"size":null,"temperature":4369.108948562811},{"position":{"x":103.89298311751936,"y":-33.38335698018097,"z":-21.8960302875648},"size":null,"temperature":8702.901602584358},{"position":{"x":-24.252910275475333,"y":-17.323625288203655,"z":133.35459012396046},"size":null,"temperature":1861.544636945028},{"position":{"x":149.8260399495654,"y":16.154518305803542,"z":48.316843714833134},"size":null,"temperature":9557.725203483238},{"position":{"x":-3.460055393107723,"y":-32.07950490930177,"z":-172.1245055894841},"size":null,"temperature":6430.856406889323},{"position":{"x":-30.220486843212676,"y":-6.287137215345295,"z":15.58193044527404},"size":null,"temperature":1230.0575488386944},{"position":{"x":-42.05825847812662,"y":-59.71778473247629,"z":6.936022760165994},"size":null,"temperature":2476.56626276512},{"position":{"x":-22.921898899429102,"y":-17.702619046205644,"z":-50.850239318345736},"size":null,"temperature":1354.8681425651853},{"position":{"x":-74.67126686085304,"y":-23.63992431950835,"z":-62.134610477912005},"size":null,"temperature":3860.972889634302},{"position":{"x":153.682928957621,"y":85.45419888691357,"z":115.87109016122388},"size":null,"temperature":5171.80895673661},{"position":{"x":-12.877117209601384,"y":-98.22175560485307,"z":-109.61058288069815},"size":null,"temperature":7136.595138877907},{"position":{"x":68.04528875159531,"y":-134.97640109811726,"z":42.31102971728972},"size":null,"temperature":7103.371026508217},{"position":{"x":-53.95710886248812,"y":75.01806473396853,"z":77.15032101866285},"size":null,"temperature":6326.26365792298},{"position":{"x":97.81178652848754,"y":-96.54342120348696,"z":55.9222049366184},"size":null,"temperature":5670.302689387605},{"position":{"x":-19.31843985699153,"y":64.92047413997385,"z":-2.865342967819066},"size":null,"temperature":4828.473951122013},{"position":{"x":34.19828074675969,"y":-74.99399315499423,"z":112.04095910400018},"size":null,"temperature":7217.978815649625},{"position":{"x":-6.990439349273185,"y":46.44453292682166,"z":-66.52801981285833},"size":null,"temperature":2018.041523647514},{"position":{"x":-114.67478238567662,"y":-33.70467673808515,"z":-133.96992443898023},"size":null,"temperature":4983.338182784308},{"position":{"x":-49.656141276766796,"y":-35.97915714561633,"z":89.29448120457886},"size":null,"temperature":7503.238301958534},{"position":{"x":68.4549641787485,"y":66.17246916630482,"z":49.124664111233656},"size":null,"temperature":11246.632092747199},{"position":{"x":162.50839128363185,"y":55.621566500834184,"z":-122.09783708064529},"size":null,"temperature":4953.715899937653},{"position":{"x":-52.33145891279182,"y":-126.76977759441382,"z":65.8750702954261},"size":null,"temperature":1950.8940721633348},{"position":{"x":132.67815785478675,"y":115.00976320336652,"z":68.03323660565324},"size":null,"temperature":6988.6042214877},{"position":{"x":-28.444848061514286,"y":48.401234991801445,"z":-60.07795204111157},"size":null,"temperature":4980.570914214743},{"position":{"x":22.020277561341207,"y":-113.61663717320074,"z":-88.20539726332112},"size":null,"temperature":1288.6824320483406},{"position":{"x":60.16801410213024,"y":-42.25808806186239,"z":82.29463159469486},"size":null,"temperature":9203.703571671482},{"position":{"x":-42.58253462875178,"y":-2.621892693146361,"z":128.90355020779603},"size":null,"temperature":6410.274056908803},{"position":{"x":17.22276751421303,"y":16.41024424319252,"z":-66.08915648140972},"size":null,"temperature":11375.58416387792},{"position":{"x":26.893146275468638,"y":27.023985685095575,"z":71.34711252277809},"size":null,"temperature":5470.738305929462},{"position":{"x":102.50075864006442,"y":-23.83534179841567,"z":-103.11288916558357},"size":null,"temperature":3364.551030734811},{"position":{"x":-58.53273820547225,"y":-207.0957672349888,"z":25.597271584352402},"size":null,"temperature":4131.657470078979},{"position":{"x":57.03710133445651,"y":-133.45399185873882,"z":15.1446185793983},"size":null,"temperature":5224.572444904863},{"position":{"x":50.93725046348107,"y":114.43697096551851,"z":-45.43430877780059},"size":null,"temperature":8145.1672996139005},{"position":{"x":-6.351695013311259,"y":28.82095708498618,"z":27.83303763742247},"size":null,"temperature":3382.2035004301943},{"position":{"x":-0.19416919179274394,"y":21.51563601345846,"z":31.822829712377434},"size":null,"temperature":1556.8953107841758},{"position":{"x":65.89661591461497,"y":32.09410938817878,"z":131.0021392061874},"size":null,"temperature":6565.054821579277},{"position":{"x":19.35738019516844,"y":-34.91478018505129,"z":74.33349999978292},"size":null,"temperature":5340.418636491717},{"position":{"x":73.57179708176561,"y":-72.7342006791038,"z":-51.564527472495456},"size":null,"temperature":1778.3058536137949},{"position":{"x":-49.961956084054336,"y":-63.41645048681639,"z":-67.22650690649937},"size":null,"temperature":6138.877718774079},{"position":{"x":123.60599223293455,"y":-80.55739985751636,"z":28.497808047945206},"size":null,"temperature":3473.443219658659},{"position":{"x":-26.274561544561827,"y":44.07576599447509,"z":-35.084942969845855},"size":null,"temperature":5952.779434133683},{"position":{"x":87.35188547018083,"y":150.6110298384523,"z":132.9450709658198},"size":null,"temperature":6164.179949166339},{"position":{"x":233.7985056544717,"y":-73.89835138773746,"z":-23.163044930985635},"size":null,"temperature":6000.734271947636},{"position":{"x":13.643724152607946,"y":-96.13868593528333,"z":4.374894183503148},"size":null,"temperature":4527.588230803418},{"position":{"x":17.65324407337288,"y":-45.58545214417578,"z":-103.27943976681635},"size":null,"temperature":5899.901133449702},{"position":{"x":-12.559801858694371,"y":-24.778861926315894,"z":143.3710513776789},"size":null,"temperature":5235.7157968151005},{"position":{"x":-41.12113707263154,"y":-133.3641922343153,"z":15.936637045431382},"size":null,"temperature":5184.895295270204},{"position":{"x":-101.10225983032149,"y":99.75191030844059,"z":-80.66524237372344},"size":null,"temperature":9992.093403819988},{"position":{"x":35.744775140958176,"y":19.076678312001203,"z":82.03415524243815},"size":null,"temperature":5954.374877714727},{"position":{"x":162.02654890895974,"y":74.31739944927152,"z":18.95792964868356},"size":null,"temperature":1724.6490724965229},{"position":{"x":93.34583778233844,"y":-47.16809338551122,"z":-48.02532005636455},"size":null,"temperature":1276.2930683215582},{"position":{"x":-147.76845661476057,"y":-100.74500681053891,"z":-37.885202110765135},"size":null,"temperature":4902.070771391536},{"position":{"x":-99.64597711675594,"y":-36.89228535858913,"z":-99.81754061804679},"size":null,"temperature":1260.3978003656482},{"position":{"x":27.37639025601696,"y":-101.49906681837255,"z":-47.771251877000466},"size":null,"temperature":14182.01033919212},{"position":{"x":33.423113158884135,"y":-41.445536239482166,"z":-70.36699950751911},"size":null,"temperature":1214.8935614223096},{"position":{"x":15.975204826179706,"y":126.43344721684457,"z":3.393613213537328},"size":null,"temperature":1414.4940848529777},{"position":{"x":-9.990102471979379,"y":-33.2694987069683,"z":-52.61727172183391},"size":null,"temperature":3692.73715871048},{"position":{"x":11.258498807558613,"y":-37.75659311608586,"z":-0.4489699966558298},"size":null,"temperature":6769.949124553216},{"position":{"x":45.09039858849278,"y":116.21329661686303,"z":18.145012876980363},"size":null,"temperature":3160.9462528307677},{"position":{"x":104.65248435592525,"y":-55.572265931896645,"z":-116.35925826832894},"size":null,"temperature":9944.015889868147},{"position":{"x":-12.438929575406668,"y":77.35899127829532,"z":61.0369868035115},"size":null,"temperature":2172.2023152616816},{"position":{"x":0.11684023391504361,"y":69.22972515904756,"z":12.813845261178253},"size":null,"temperature":9528.895711772559},{"position":{"x":-35.03187439913218,"y":20.12368998560722,"z":33.2212506349426},"size":null,"temperature":4057.418684036199},{"position":{"x":61.014930696052545,"y":-16.566091385629516,"z":-150.7240341303331},"size":null,"temperature":3478.889459035773}]'
  );

  $.getJSON("http://localhost:8080/galaxy", function(dataN) {
    $.each(dataN, function(k1, star) {
      console.log(star);
      stars.push({
        x: star.X,
        y: star.Y,
        z: star.Z
      });

      starsColor.push(new THREE.Color(star.Color)); //0xff0000*/

      $(".Slist").append(
        '<li data-seed="' +
          star.ID +
          '" data-posX="' +
          star.X +
          '" data-posY="' +
          star.Y +
          '" data-posZ="' +
          star.Z +
          '">' +
          star.Name +
          " - " +
          star.ID +
          "</li>"
      );
    });

    setGalaxy();
    // updateInfos();
    // updateControls();
  });

  // setGalaxy();

  // var url =
  //   "https://api.global-chaos.fr/api/galaxy?nbStar=500&galaxySize=" +
  //   galaxySize +
  //   "&galaxyArm=" +
  //   galaxyArm +
  //   "&galaxyRoll=" +
  //   galaxyRoll;

  // if (nextSeed) url = url + "&seed=" + nextSeed;

  //   $.getJSON(url, function(dataN) {
  //     var tmpList = [];
  //     if (!nextSeed) firstSeed = dataN.seed;
  //     $.each(dataN.systems, function(key, val) {
  //       nbStar++;
  //       // starList.push(val);
  //       // tmpList.push(val);
  //       stars.push({
  //         x: val.coord.x,
  //         y: val.coord.y,
  //         z: val.coord.z
  //       });
  //       starsColor.push(
  //         new THREE.Color(
  //           parseInt(
  //             "0x" +
  //               val.color.r.toString(16) +
  //               val.color.g.toString(16) +
  //               val.color.b.toString(16),
  //             16
  //           )
  //         )
  //       ); //0xff0000*/
  //       if (val.planets.length) {
  //         if (val.habitable) {
  //           nbHS++;
  //           nbHP += val.habitable;
  //         } else {
  //           nbNHP += val.planets.length - val.habitable;
  //         }
  //       }
  //       /*else
  //                         starsColor.push(new THREE.Color(0xff0000)); //0xff0000*/
  //       $(".Slist").append(
  //         '<li data-seed="' +
  //           val.seed +
  //           '" data-posX="' +
  //           val.coord.x +
  //           '" data-posY="' +
  //           val.coord.y +
  //           '" data-posZ="' +
  //           val.coord.z +
  //           '">' +
  //           val.name +
  //           "</li>"
  //       );
  //     });
  //     setGalaxy();
  //     updateInfos();
  //     updateControls();
  //     if (nbStar <= maxStar) streamStar(dataN.nextseed);
  //     else {
  //       $(".Slist li")
  //         .sort(function(a, b) {
  //           return $(b).text() < $(a).text() ? 1 : -1;
  //         })
  //         .appendTo(".Slist");
  //     }
  //   });
}

function buildAxis(src, dst, colorHex) {
  var geom = new THREE.Geometry(),
    mat;

  mat = new THREE.LineBasicMaterial({ linewidth: 1, color: colorHex });
  geom.vertices.push(src.clone());
  geom.vertices.push(dst.clone());

  var axis = new THREE.Line(geom, mat, THREE.LineSegments);

  return axis;
}

function buildAxes(pos, length) {
  var axes = new THREE.Object3D();

  axes.add(
    buildAxis(
      pos.clone(),
      new THREE.Vector3(length + pos.x, 0 + pos.y, 0 + pos.z),
      0x008080
    )
  ); // +X
  axes.add(
    buildAxis(
      pos.clone(),
      new THREE.Vector3(-length + pos.x, 0 + pos.y, 0 + pos.z),
      0x008080
    )
  ); // -X
  axes.add(
    buildAxis(
      pos.clone(),
      new THREE.Vector3(0 + pos.x, length + pos.y, 0 + pos.z),
      0x008080
    )
  ); // +Y
  axes.add(
    buildAxis(
      pos.clone(),
      new THREE.Vector3(0 + pos.x, -length + pos.y, 0 + pos.z),
      0x008080
    )
  ); // -Y
  axes.add(
    buildAxis(
      pos.clone(),
      new THREE.Vector3(0 + pos.x, 0 + pos.y, length + pos.z),
      0x008080
    )
  ); // +Z
  axes.add(
    buildAxis(
      pos.clone(),
      new THREE.Vector3(0 + pos.x, 0 + pos.y, -length + pos.z),
      0x008080
    )
  ); // -Z

  return axes;
}

//threejs functions
function setScene() {
  scene = new THREE.Scene();

  camera = new THREE.PerspectiveCamera(
    45,
    innerWidth / innerHeight,
    0.5,
    50000
  );
  camera.position.set(1500, 1500, 1500);
  //   camera.position.set(0, 0, 0);

  renderer = new THREE.WebGLRenderer();
  renderer.setSize(innerWidth, innerHeight);
  renderer.setClearColor(0x00000000);
  document.body.appendChild(renderer.domElement);

  controls = new THREE.TrackballControls(camera, renderer.domElement);
  controls.noPan = true;
  controls.noZoom = false;
  controls.dynamicDampingFactor = 0.1;
  controls.rotateSpeed = 5;

  setGalaxy();

  raycaster = new THREE.Raycaster();
  mouse = new THREE.Vector2();

  // Event Listener
  document.addEventListener("mousemove", onDocumentMouseMove, false);
  document.addEventListener("click", onDocumentMouseClick, false);
  window.addEventListener("resize", onWindowResize, false);
}

function onWindowResize() {
  camera.aspect = window.innerWidth / window.innerHeight;
  renderer.setSize(window.nnerWidth, window.innerHeight);
  camera.updateProjectionMatrix();
  renderer.render(scene, camera);
}

function onDocumentMouseMove(event) {
  event.preventDefault();
  mouse.x = (event.clientX / window.innerWidth) * 2 - 1;
  mouse.y = -(event.clientY / window.innerHeight) * 2 + 1;
}

function onDocumentMouseClick(event) {
  if (INTERSECTED != null) {
    if (typeof locator != "undefined") scene.remove(locator);
    locator = buildAxes(
      new THREE.Vector3(
        starList[INTERSECTED].coord.x,
        starList[INTERSECTED].coord.y,
        starList[INTERSECTED].coord.z
      ),
      100
    );
    scene.add(locator);
    controls.target.set(
      starList[INTERSECTED].coord.x,
      starList[INTERSECTED].coord.y,
      starList[INTERSECTED].coord.z
    );
    $(".Slist li").each(function() {
      $(this).removeClass("active");
    });
    $('.Slist li[data-seed="' + starList[INTERSECTED].seed + '"]').addClass(
      "active"
    );
    $(".Slist").scrollTo("li.active");
    displaySystemInfo(starList[INTERSECTED].seed);
  }
}

function findSystem(seed) {
  return $.grep(starList, function(item) {
    return item.seed == seed;
  });
}

function displaySystemInfo(seed) {
  var star = findSystem(seed)["0"];

  $("#systemInfo .Name").html(star.name);
  $("#systemInfo .SpectralClass").html(
    star.spectralType + star.spectralSubtype + star.Lclass
  );
  $("#systemInfo .Temperature").html(star.temperature);
  $("#systemInfo .Mass").html(star.mass);
  $("#systemInfo .Radius").html(star.radius);
  $("#systemInfo .Gravity").html(star.gravity);
}

function setGalaxy() {
  galaxyMaterial = new THREE.ShaderMaterial({
    vertexShader: document.getElementById("vShader").textContent,
    vertexColors: THREE.VertexColors,
    fragmentShader: document.getElementById("fShader").textContent,
    uniforms: {
      size: {
        type: "f",
        value: 10
      }, //10
      t: {
        type: "f",
        value: 0
      },
      z: {
        type: "f",
        value: 0
      },
      pixelRatio: {
        type: "f",
        value: innerHeight
      }
    },
    transparent: true,
    depthTest: false,
    blending: THREE.AdditiveBlending
  });
  var geometry = new THREE.Geometry();
  geometry.vertices = stars;
  geometry.colors = starsColor;
  geometry.colorsNeedUpdate = true;

  if (typeof galaxy != "undefined") scene.remove(galaxy);

  galaxy = new THREE.Points(geometry, galaxyMaterial);
  scene.add(galaxy);
}

function animate() {
  runLoopIdentifier = requestAnimationFrame(animate);

  raycaster.setFromCamera(mouse, camera);
  var intersects = raycaster.intersectObject(galaxy);

  if (intersects.length > 0) {
    if (INTERSECTED != intersects["0"].index) {
      if (INTERSECTED) console.log("OUT");
      INTERSECTED = intersects["0"].index;
      //console.log(starList[ intersects['0'].index ]);
    }
  } else {
    if (INTERSECTED) console.log("OUT");
    INTERSECTED = null;
  }

  renderer.render(scene, camera);
  //scene.rotation.z -= 0.001;
  controls.update();
  //stats.update();
}