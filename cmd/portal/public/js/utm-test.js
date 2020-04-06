// UTM CONVERTION
function validate(latitude, longitude)
{
    if (latitude < -90.0 || latitude > 90.0 || longitude < -180.0
        || longitude >= 180.0)
    {
    throw new IllegalArgumentException(
        "Legal ranges: latitude [-90,90], longitude [-180,180).");
    }
}

function degreeToRadian(degree)
{
    return degree * Math.PI / 180;
}

function radianToDegree(radian)
{
    return radian * 180 / Math.PI;
}

function POW(a, b)
{
    return Math.pow(a, b);
}

function SIN(value)
{
    return Math.sin(value);
}

function COS(value)
{
    return Math.cos(value);
}

function TAN(value)
{
    return Math.tan(value);
}

function convertLatLonToUTM(latitude, longitude)
{
    // Lat Lon to UTM variables

    debugger;

    // equatorial radius
    var equatorialRadius = 6378137;

    // polar radius
    var polarRadius = 6356752.314;

    // scale factor
    var k0 = 0.9996;

    // eccentricity
    var e = Math.sqrt(1 - POW(polarRadius / equatorialRadius, 2));

    var e1sq = e * e / (1 - e * e);

    var n = (equatorialRadius - polarRadius)
        / (equatorialRadius + polarRadius);

    // r curv 1
    var rho = 6368573.744;

    // r curv 2
    var nu = 6389236.914;

    // Calculate Meridional Arc Length
    // Meridional Arc
    var S = 5103266.421;

    var A0 = 6367449.146;

    var B0 = 16038.42955;

    var C0 = 16.83261333;

    var D0 = 0.021984404;

    var E0 = 0.000312705;

    // Calculation Constants
    // Delta Long
    var p = -0.483084;

    var sin1 = 4.84814E-06;

    // Coefficients for UTM Coordinates
    var K1 = 5101225.115;

    var K2 = 3750.291596;

    var K3 = 1.397608151;

    var K4 = 214839.3105;

    var K5 = -2.995382942;

    var A6 = -1.00541E-07;

    if (latitude < -90.0 || latitude > 90.0 || longitude < -180.0
        || longitude >= 180.0)
    {
        throw new IllegalArgumentException(
            "Legal ranges: latitude [-90,90], longitude [-180,180)."
        );
    }
    let UTM = "";

    //var latitude = degreeToRadian(latitude);
    rho = equatorialRadius * (1 - e * e)
        / POW(1 - POW(e * SIN(latitude), 2), 3 / 2.0);

    nu = equatorialRadius / POW(1 - POW(e * SIN(latitude), 2), (1 / 2.0));

    var var1;
    if (longitude < 0.0)
    {
        var1 = (((180 + longitude) / 6.0)) + 1;
    }
    else
    {
        var1 = (longitude / 6) + 31;
    }
    var var2 = (6 * var1) - 183;
    var var3 = longitude - var2;
    p = var3 * 3600 / 10000;

    S = A0 * latitude - B0 * SIN(2 * latitude) + C0 * SIN(4 * latitude) - D0
        * SIN(6 * latitude) + E0 * SIN(8 * latitude);

    K1 = S * k0;
    K2 = nu * SIN(latitude) * COS(latitude) * POW(sin1, 2) * k0 * (100000000)
        / 2;
    K3 = ((POW(sin1, 4) * nu * SIN(latitude) * Math.pow(COS(latitude), 3)) / 24)
        * (5 - POW(TAN(latitude), 2) + 9 * e1sq * POW(COS(latitude), 2) + 4
            * POW(e1sq, 2) * POW(COS(latitude), 4))
        * k0
        * 10000000000000000;

    K4 = nu * COS(latitude) * sin1 * k0 * 10000;

    K5 = POW(sin1 * COS(latitude), 3) * (nu / 6)
        * (1 - POW(TAN(latitude), 2) + e1sq * POW(COS(latitude), 2)) * k0
        * 1000000000000;

    A6 = (POW(p * sin1, 6) * nu * SIN(latitude) * POW(COS(latitude), 5) / 720)
        * (61 - 58 * POW(TAN(latitude), 2) + POW(TAN(latitude), 4) + 270
            * e1sq * POW(COS(latitude), 2) - 330 * e1sq
            * POW(SIN(latitude), 2)) * k0 * (1E+24);

    debugger;
    var longZone = getLongZone(longitude);
    var latZone = getLatZone(latitude);

    var _easting = getEasting();
    var _northing = getNorthing(latitude);

    UTM = longZone + " " + latZone + " " + _easting + " "
        + _northing;
    // UTM = longZone + " " + latZone + " " + decimalFormat.format(_easting) +
    // " "+ decimalFormat.format(_northing);

    return UTM;

}

function setVariables(latitude, longitude)
{
    //var latitude = degreeToRadian(latitude);
    var rho = equatorialRadius * (1 - e * e)
        / POW(1 - POW(e * SIN(latitude), 2), 3 / 2.0);

    var nu = equatorialRadius / POW(1 - POW(e * SIN(latitude), 2), (1 / 2.0));

    var var1;
    if (longitude < 0.0)
    {
        var1 = ((int) ((180 + longitude) / 6.0)) + 1;
    }
    else
    {
        var1 = ((int) (longitude / 6)) + 31;
    }
    var var2 = (6 * var1) - 183;
    var var3 = longitude - var2;
    p = var3 * 3600 / 10000;

    S = A0 * latitude - B0 * SIN(2 * latitude) + C0 * SIN(4 * latitude) - D0
        * SIN(6 * latitude) + E0 * SIN(8 * latitude);

    K1 = S * k0;
    K2 = nu * SIN(latitude) * COS(latitude) * POW(sin1, 2) * k0 * (100000000)
        / 2;
    K3 = ((POW(sin1, 4) * nu * SIN(latitude) * Math.pow(COS(latitude), 3)) / 24)
        * (5 - POW(TAN(latitude), 2) + 9 * e1sq * POW(COS(latitude), 2) + 4
            * POW(e1sq, 2) * POW(COS(latitude), 4))
        * k0
        * 10000000000000000;

    K4 = nu * COS(latitude) * sin1 * k0 * 10000;

    K5 = POW(sin1 * COS(latitude), 3) * (nu / 6)
        * (1 - POW(TAN(latitude), 2) + e1sq * POW(COS(latitude), 2)) * k0
        * 1000000000000;

    A6 = (POW(p * sin1, 6) * nu * SIN(latitude) * POW(COS(latitude), 5) / 720)
        * (61 - 58 * POW(TAN(latitude), 2) + POW(TAN(latitude), 4) + 270
            * e1sq * POW(COS(latitude), 2) - 330 * e1sq
            * POW(SIN(latitude), 2)) * k0 * (1E+24);
}

function getLongZone(longitude)
{
    let longZone = 0;
    if (longitude < 0.0)
    {
        longZone = ((180.0 + longitude) / 6) + 1;
    }
    else
    {
        longZone = (longitude / 6) + 31;
    }
    let val = String.valueOf(longZone);
    if (val.length() == 1)
    {
        val = "0" + val;
    }
    return val;
}

function getLatZone(latitude)
{
    let latIndex = -2;
    let lat = latitude;
    let posLetters = [
        'N',
        'P',
        'Q',
        'R',
        'S',
        'T',
        'U',
        'V',
        'W',
        'X',
        'Z'
    ];
    let posDegrees = [
        0,
        8,
        16,
        24,
        32,
        40,
        48,
        56,
        64,
        72,
        84
    ];
    let negLetters = [
        'A',
        'C',
        'D',
        'E',
        'F',
        'G',
        'H',
        'J',
        'K',
        'L',
        'M'
    ];

    if (lat >= 0)
    {
        posLetters.some((letter, i) => {
            if (lat == letter)
            {
                latIndex = i;
                return;
            }
            else if (lat < posDegrees[i])
            {
                latIndex = i - 1;
                return;
            }
        });
    }
    else
    {
        let len = negLetters.length;
        negLetters.forEach((letter, i) => {
            if (lat == negDegrees[i])
            {
                latIndex = i;
                return;
            }
            if (lat < negDegrees[i])
            {
                latIndex = i - 1;
                return;
            }
        });
    }

    if (latIndex == -1)
    {
        latIndex = 0;
    }
    if (lat >= 0)
    {
        if (latIndex == -2)
        {
            latIndex = posLetters.length - 1;
        }
        return String.valueOf(posLetters[latIndex]);
    }
    else
    {
        if (latIndex == -2)
        {
            latIndex = negLetters.length - 1;
        }
        return String.valueOf(negLetters[latIndex]);
    }
}
