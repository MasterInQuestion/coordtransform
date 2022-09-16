/*

= "coordTransform" =

[
|*| Source: https://github.com/MasterInQuestion/coordtransform/raw/master/main.go
|*| Last update: CE 2022-09-16 21:18 UTC ]


Go implementation for converting several China obfuscated GPS coordinate schemas back into the regular form (and vice versa).

Currently supported:
|*| WGS84 <-> GCJ02
|*| WGS84 <-> BD09 ("bd09ll")
|*| WGS84 <-> BD09MC
|*| GCJ02 <-> BD09 ("bd09ll")
|*| GCJ02 <-> BD09MC
|*| BD09 ("bd09ll") <-> BD09MC

*/



// == Implementation ==

//	package coordTransform;
	package main;

	import (
	. `fmt`;
	. `math`;
	`regexp`;
	);


	const (
	EnableBorderDetection bool = false;

	Axis float64 = 6378245;
	Offset float64 = 0.006693421622965943;
/*

Offset =
((
2 / F - 1 / F^2
))
, where:
|*| F = 298.3

*/
	Pi_180 float64 = Pi / 180;
	);


// === Subroutines ===

	func inChina (
	N float64,
	E float64,
	) (
	bool,
	) {

	{ return (
// Vastly off: apparently more than in China.
	N >= 17.95752 &&
	N <= 53.56082 &&

	E >= 73.55 &&
	E <= 134.75 )
	};
/*

Region preview:
|*| https://www.openstreetmap.org/?bbox=73.55,17.95752,134.75,53.56082&mlat=17.95752&mlon=73.55
|*| https://www.openstreetmap.org/?bbox=73.55,17.95752,134.75,53.56082&mlat=53.56082&mlon=134.75


But no matter how the function is designed it would always fail on border cases. (per the incompatible nature of GCJ02)

*/
	};


// === Primary functions ===
/*

[
|*| WGS84: 又 地球坐标系, 国际通行坐标系.
|*| GCJ02: 又 火星坐标系, 由 WGS84 混淆后的坐标系. Google Maps, 高德 在用.
|*| BD09 ("bd09ll"): 又 百度坐标系, 由 GCJ02 混淆后的坐标系. 应用于部分 百度地图 API. ]

|*| WGS84: known as Coordinate System for Earth, the internationally exchangeable coordinate system.
|*| GCJ02: known as Coordinate System for Mars, obfuscated coordinate system based on WGS84. Used by Google Maps, Amap (高德).
|*| BD09 ("bd09ll"): known as Baidu Coordinate System, obfuscated coordinate system based on GCJ02. Used by some Baidu Map's API.
|*| BD09MC: Baidu's another vain attempt at white-box cryptography. The schema's output sort of resembles EPSG3857 [ https://epsg.io/3857 ]. Used by some Baidu Map's API.

*/
// ==== Basic ====

// WGS84 -> GCJ02:
	func WGS84toGCJ02 (
	N float64,
	E float64,
	) (
	float64,
	float64,
	) {
/*

Current implementation of this function provides suboptimal accuracy (~ 1 m).


Several other implementations have been referred:
|*| https://github.com/wandergis/coordtransform
|*| https://github.com/kikkimo/WgsToGcj/blob/master/src/WGS2GCJ.cpp
|*| https://github.com/Artoria2e5/PRCoords
|*| https://chaoli.club/index.php/4777/p1#p49268
|*| https://atool.vip/lnglat/

; but only find out that they were either no better or worse.


[ Additional Note:

This implementation appears to give output practically identical to:
|*| Amap's API ([ https://uri.amap.com/marker?coordinate=wgs84&position=E,N ], replace the "position" parameter accordingly; or try this online demo: [ https://lbs.amap.com/api/webservice/guide/api/convert#satisfy-container ]);
|*| Tencent Map's API ([ https://apis.map.qq.com/uri/v1/marker?coord_type=1&marker=title:-;coord:N,E ], adapt the "coord" part of the "marker" parameter accordingly).

~~The 2 services seem to interpret the coordinates specially: the very "incorrect" coordinates would map to acceptable locations in their interface whereas the "correct" ones would be off.~~ [ Conclusion based on false premise. ] ]

*/
	if (
	EnableBorderDetection &&
	! inChina( N, E ) ) {

	return N, E;
	};


	_N := N - 35;
	_E := E - 105;

	x0 := _N * Pi;
	x1 := _E * Pi;
	x2 := _N * _E;
	x3 := Sqrt( Abs( _E ) );
	x4 := 20 * (Sin( x1 * 6 ) + Sin( x1 * 2 ) );

	x5 := N * Pi_180;
	x6 := 1 - Offset * Pow( Sin( x5 ), 2 );
	x7 := Axis / Sqrt( x6 );

	{ return (
	N + ( (x4 +
	20 * Sin( x0 ) +
	40 * Sin( x0 / 3 ) +
	160 * Sin( x0 / 12 ) +
	320 * Sin( x0 / 30 ) ) / 1.5 +
	_N * 3 + _E * 2 + x2 / 10 + (Pow( _N, 2 ) + x3) / 5 - 100) /
	x7 / (1 - Offset) * x6 / Pi_180 ),

	(
	E + ( (x4 +
	20 * Sin( x1 ) +
	40 * Sin( x1 / 3 ) +
	150 * Sin( x1 / 12 ) +
	300 * Sin( x1 / 30 ) ) / 1.5 +
	_N * 2 + _E + (Pow( _E, 2 ) + x2 + x3) / 10 + 300) /
	x7 / Cos( x5 ) / Pi_180 )
	};

	};


// WGS84 -> BD09 ("bd09ll"):
	func WGS84toBD09 (
	N float64,
	E float64,
	) (
	float64,
	float64,
	) {

	return GCJ02toBD09( WGS84toGCJ02( N, E ) );
	};


// WGS84 -> BD09MC:
	func WGS84toBD09MC (
	N float64,
	E float64,
	) (
	float64,
	float64,
	) {

	return BD09toBD09MC( WGS84toBD09( N, E ) );
	};


// GCJ02 -> WGS84:
	func GCJ02toWGS84 (
	N float64,
	E float64,
	) (
	float64,
	float64,
	) {

	if (
	EnableBorderDetection &&
	! inChina( N, E ) ) {

	return N, E;
	};


	_N := N - 35;
	_E := E - 105;

	x0 := _N * Pi;
	x1 := _E * Pi;
	x2 := _N * _E;
	x3 := Sqrt( Abs( _E ) );
	x4 := 20 * (Sin( x1 * 6 ) + Sin( x1 * 2 ) );

	x5 := N * Pi_180;
	x6 := 1 - Offset * Pow( Sin( x5 ), 2 );
	x7 := Axis / Sqrt( x6 );

	{ return (
	N - ( (x4 +
	20 * Sin( x0 ) +
	40 * Sin( x0 / 3 ) +
	160 * Sin( x0 / 12 ) +
	320 * Sin( x0 / 30 ) ) / 1.5 +
	_N * 3 + _E * 2 + x2 / 10 + (Pow( _N, 2 ) + x3) / 5 - 100) /
	x7 / (1 - Offset) * x6 / Pi_180 ),

	(
	E - ( (x4 +
	20 * Sin( x1 ) +
	40 * Sin( x1 / 3 ) +
	150 * Sin( x1 / 12 ) +
	300 * Sin( x1 / 30 ) ) / 1.5 +
	_N * 2 + _E + (Pow( _E, 2 ) + x2 + x3) / 10 + 300) /
	x7 / Cos( x5 ) / Pi_180 )
	};

	};


// BD09 ("bd09ll") -> WGS84:
	func BD09toWGS84 (
	N float64,
	E float64,
	) (
	float64,
	float64,
	) {

	return GCJ02toWGS84( BD09toGCJ02( N, E ) );
	};


// BD09MC -> WGS84:
	func BD09MCtoWGS84 (
	X float64,
	Y float64,
	) (
	float64,
	float64,
	) {

	return BD09toWGS84( BD09MCtoBD09( X, Y ) );
	};

// ==== Extended ====

// ===== BD09 ("bd09ll") =====

	const x_Pi float64 = Pi_180 * 3000;


// GCJ02 -> BD09 ("bd09ll"):
	func GCJ02toBD09 (
	N float64,
	E float64,
	) (
	float64,
	float64,
	) {

	x0 := Sqrt( Pow( N, 2 ) + Pow( E, 2 ) ) + 0.00002 * Sin( N * x_Pi );
	x1 := Atan2( N, E ) + 0.000003 * Cos( E * x_Pi );

	{ return (
	x0 * Sin( x1 ) + 0.006 ),
	(
	x0 * Cos( x1 ) + 0.0065 )
	};

	};


// BD09 ("bd09ll") -> GCJ02:
	func BD09toGCJ02 (
	N float64,
	E float64,
	) (
	float64,
	float64,
	) {

	N -= 0.006;
	E -= 0.0065;

	x0 := Sqrt( Pow( N, 2 ) + Pow( E, 2 ) ) - 0.00002 * Sin( N * x_Pi );
	x1 := Atan2( N, E ) - 0.000003 * Cos( E * x_Pi );

	{ return (
	x0 * Sin( x1 ) ),
	(
	x0 * Cos( x1 ) )
	};

	};

// ===== BD09MC =====

// BD09 ("bd09ll") -> BD09MC:
	func BD09toBD09MC (
	N float64,
	E float64,
	) (
	X float64,
	Y float64,
	) {
/*

See "BD09MCtoBD09" for notes.

*/
	if ( E > 180 ) {
	E -= 360 * Floor( (E + 180) / 360 );

	} else if ( E < -180 ) {
	E -= 360 * Ceil( (E - 180) / 360 );
	};


	n := [10](float64){};
	x0 := Abs( N );

	if ( x0 >= 60 ) {
// Suboptimal accuracy.
	n[0] = 0.0008277824516172526;
	n[1] = 111320.7020463578;
	n[2] = 647795574.6671607;
	n[3] = -4082003173.641316;
	n[4] = 10774905663.51142;
	n[5] = -15171875531.51559;
	n[6] = 12053065338.62167;
	n[7] = -5124939663.577472;
	n[8] = 913311935.9512032;
	n[9] = 67.5;

	} else if ( x0 >= 45 ) {
// Suboptimal accuracy.
	n[0] = 0.00337398766765;
	n[1] = 111320.7020202162;
	n[2] = 4481351.045890365;
	n[3] = -23393751.19931662;
	n[4] = 79682215.47186455;
	n[5] = -115964993.2797253;
	n[6] = 97236711.15602145;
	n[7] = -43661946.33752821;
	n[8] = 8477230.501135234;
	n[9] = 52.5;

	} else if ( x0 >= 30 ) {
	n[0] = 0.00220636496208;
	n[1] = 111320.7020209128;
	n[2] = 51751.86112841131;
	n[3] = 3796837.749470245;
	n[4] = 992013.7397791013;
	n[5] = -1221952.21711287;
	n[6] = 1340652.697009075;
	n[7] = -620943.6990984312;
	n[8] = 144416.9293806241;
	n[9] = 37.5;

	} else if ( x0 >= 15 ) {
	n[0] = -0.0003441963504368392;
	n[1] = 111320.7020576856;
	n[2] = 278.2353980772752;
	n[3] = 2485758.690035394;
	n[4] = 6070.750963243378;
	n[5] = 54821.18345352118;
	n[6] = 9540.606633304236;
	n[7] = -2710.55326746645;
	n[8] = 1405.483844121726;
	n[9] = 22.5;

	} else {
	n[0] = -0.0003218135878613132;
	n[1] = 111320.7020701615;
	n[2] = 0.00369383431289;
	n[3] = 823725.6402795718;
	n[4] = 0.46104986909093;
	n[5] = 2351.343141331292;
	n[6] = 1.58060784298199;
	n[7] = 8.77738589078284;
	n[8] = 0.37238884252424;
	n[9] = 7.45;
	};

// X:
	{
	X = n[0] + n[1] * Abs( E );
	if ( E < 0 ) { X *= -1; };
	};

// Y:
	{
	x0 /= n[9];
	x1 := Pow( x0, 2 );

	Y = n[2] + n[3] * x0 + n[4] * x1;

	x1 *= x0;
	Y += n[5] * x1;
	x1 *= x0;
	Y += n[6] * x1;
	x1 *= x0;
	Y += (n[7] + n[8] * x0) * x1;

	if ( N < 0 ) { Y *= -1; };


	if ( Y > 19429903 ) {
	Y = 19429903;

	} else if ( Y < -16022031 ) {
	Y = -16022031;
	};

	};

	return;
	};


// GCJ02 -> BD09MC:
	func GCJ02toBD09MC (
	N float64,
	E float64,
	) (
	float64,
	float64,
	) {

	return BD09toBD09MC( GCJ02toBD09( N, E ) );
	};


// BD09MC -> BD09 ("bd09ll"):
	func BD09MCtoBD09 (
	X float64,
	Y float64,
	) (
	N float64,
	E float64,
	) {
/*

Both this and the "BD09toBD09MC" function have been verified of being algorithmically correct. But some of the interpolation modifiers (values conditionally assigned to "n") are either incorrect or of inadequate accuracy.

*/
	if ( X > 20037726.372307256 ) {
	X -= 40075452.744614512 * Floor( (X + 20037726.372307256) / 40075452.744614512 );

	} else if ( X < -20037726.372307256 ) {
	X -= 40075452.744614512 * Ceil( (X - 20037726.372307256) / 40075452.744614512 );
	};


	if ( Y > 19429903 ) {
	Y = 19429903;

	} else if ( Y < -16022031 ) {
	Y = -16022031;
	};


	n := [10](float64){};
	x0 := Abs( Y );

	if ( x0 >= 8362377.87 ) {
// Failed accuracy.
	n[0] = -0.000000007435856389565537;
	n[1] = 0.000008983055097726239;
	n[2] = -0.78625201886289;
	n[3] = 96.32687599759846;
	n[4] = -1.85204757529826;
	n[5] = -59.36935905485877;
	n[6] = 47.40033549296737;
	n[7] = -16.50741931063887;
	n[8] = 2.28786674699375;
	n[9] = 10260144.86;

	} else if ( x0 >= 5591021 ) {
// Slight accuracy off.
	n[0] = -0.00000003030883460898826;
	n[1] = 0.00000898305509983578;
	n[2] = 0.30071316287616;
	n[3] = 59.74293618442277;
	n[4] = 7.357984074871;
	n[5] = -25.38371002664745;
	n[6] = 13.45380521110908;
	n[7] = -3.29883767235584;
	n[8] = 0.32710905363475;
	n[9] = 6856817.37;

	} else if ( x0 >= 3481989.83 ) {
	n[0] = -0.00000001981981304930552;
	n[1] = 0.000008983055099779535;
	n[2] = 0.03278182852591;
	n[3] = 40.31678527705744;
	n[4] = 0.65659298677277;
	n[5] = -4.44255534477492;
	n[6] = 0.85341911805263;
	n[7] = 0.12923347998204;
	n[8] = -0.04625736007561;
	n[9] = 4482777.06;

	} else if ( x0 >= 1678043.12 ) {
	n[0] = 0.00000000309191371068437;
	n[1] = 0.000008983055096812155;
	n[2] = 0.00006995724062;
	n[3] = 23.10934304144901;
	n[4] = -0.00023663490511;
	n[5] = -0.6321817810242;
	n[6] = -0.00663494467273;
	n[7] = 0.03430082397953;
	n[8] = -0.00466043876332;
	n[9] = 2555164.4;

	} else {
	n[0] = 0.000000002890871144776878;
	n[1] = 0.000008983055095805407;
	n[2] = -0.00000003068298;
	n[3] = 7.47137025468032;
	n[4] = -0.00000353937994;
	n[5] = -0.02145144861037;
	n[6] = -0.00001234426596;
	n[7] = 0.00010322952773;
	n[8] = -0.00000323890364;
	n[9] = 826088.5;
	};

// N:
	{
	x0 /= n[9];
	x1 := Pow( x0, 2 );

	N = n[2] + n[3] * x0 + n[4] * x1;

	x1 *= x0;
	N += n[5] * x1;
	x1 *= x0;
	N += n[6] * x1;
	x1 *= x0;
	N += (n[7] + n[8] * x0) * x1;

	if ( Y < 0 ) { N *= -1; };
	};

// E:
	{
	E = n[0] + n[1] * Abs( X );
	if ( X < 0 ) { E *= -1; };
	};

	return;
	};


// BD09MC -> GCJ02:
	func BD09MCtoGCJ02 (
	X float64,
	Y float64,
	) (
	float64,
	float64,
	) {

	return BD09toGCJ02( BD09MCtoBD09( X, Y ) );
	};


// === Testcase ===
/*

Try at [ https://play.golang.org/p/UCPdVNrl0-P#code ]. (slight modification to the code required)

==== Demo places ====

(40.049694° N, 116.294717° E; WGS84) China, Běi-Jīng, Seas Settled District, Top Land 10th Street #10: Baidu Building (中国, 北京, 海淀区, 上地十街 #10: 百度大厦):
|*| https://www.openstreetmap.org/?mlat=40.049694&mlon=116.294717#map=19/40.049694/116.294717
|*| https://www.google.com/maps/place/40.050962,116.30081/@40.050962,116.30081,19z?hl=en
|*| https://map.baidu.com/?latlng=40.057117,116.307236
|*| https://map.baidu.com/@12947403.2,4846489.6,19z

(22.543415° N, 113.929665° E; WGS84) China, Canton Province, Shēn-Zhèn, South Mountain District, South Shēn Avenue #10000: Tencent Building (中国, 广东省, 深圳, 南山区, 深南大道 #10000: 腾讯大厦):
|*| https://api.map.baidu.com/marker?output=html&coord_type=wgs84&location=22.543415,113.929665
|*| https://api.map.baidu.com/marker?output=html&coord_type=gcj02&location=22.540385,113.934532
|*| https://api.map.baidu.com/marker?output=html&coord_type=bd09ll&location=22.546054,113.94108
|*| https://api.map.baidu.com/marker?output=html&coord_type=bd09mc&location=12684001,2560682.4

Note "api.map.baidu.com" would response to Mainland China IP only.

*/

	func main () {
	WGS84 := [2](float64){ 22.543415, 113.929665 };
	GCJ02 := [2](float64){ 22.540385, 113.934532 };
	BD09 := [2](float64){ 22.546054, 113.94108 };
	BD09MC := [2](float64){ 12684001, 2560682.4 };

	Print(
	`|*| WGS84toGCJ02: ` + fp6( WGS84toGCJ02( WGS84[0], WGS84[1] ) ) + "\n", // Yet problematic.
	`|*| WGS84toBD09: ` + fp6( WGS84toBD09( WGS84[0], WGS84[1] ) ) + "\n",
	`|*| WGS84toBD09MC: ` + fp1( WGS84toBD09MC( WGS84[0], WGS84[1] ) ) + "\n",
	"\n",
	`|*| GCJ02toWGS84: ` + fp6( GCJ02toWGS84( GCJ02[0], GCJ02[1] ) ) + "\n",
	`|*| BD09toWGS84: ` + fp6( BD09toWGS84( BD09[0], BD09[1] ) ) + "\n",
	`|*| BD09MCtoWGS84: ` + fp6( BD09MCtoWGS84( BD09MC[0], BD09MC[1] ) ) + "\n",
	"\n",
	`|*| GCJ02toBD09: ` + fp6( GCJ02toBD09( GCJ02[0], GCJ02[1] ) ) + "\n",
	`|*| GCJ02toBD09MC: ` + fp1( GCJ02toBD09MC( GCJ02[0], GCJ02[1] ) ) + "\n",
	`|*| BD09toGCJ02: ` + fp6( BD09toGCJ02( BD09[0], BD09[1] ) ) + "\n",
	`|*| BD09MCtoGCJ02: ` + fp6( BD09MCtoGCJ02( BD09MC[0], BD09MC[1] ) ) + "\n",
	"\n",
	`|*| BD09toBD09MC: ` + fp1( BD09toBD09MC( BD09[0], BD09[1] ) ) + "\n",
	`|*| BD09MCtoBD09: ` + fp6( BD09MCtoBD09( BD09MC[0], BD09MC[1] ) ) + "\n", // Caveat minor problems.
	);

	};




	var (
	RE_0 = regexp.MustCompile( `\.?0+$` ); // Warning: This RegEx is not general-purpose safe.
	);


	func fp6 (
	_0 float64,
	_1 float64,
	) (
	string,
	) {

	{
	_0 := RE_0.ReplaceAllLiteralString( Sprintf( `%.6f`, _0 ), `` );
	_1 := RE_0.ReplaceAllLiteralString( Sprintf( `%.6f`, _1 ), `` );

	return ( _0 + `, ` + _1 );
	};

	};


	func fp1 (
	_0 float64,
	_1 float64,
	) (
	string,
	) {

	{ return (
	RE_0.ReplaceAllLiteralString( Sprintf( `%.1f`, _0 ), `` ) +
	`, ` +
	RE_0.ReplaceAllLiteralString( Sprintf( `%.1f`, _1 ), `` ) )
	};

	};
