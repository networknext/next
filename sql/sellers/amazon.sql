
-- amazon datacenters

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.johannesburg.1',
   'afs1-az1',
   -33.924900,
   18.424101,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.johannesburg.2',
   'afs1-az2',
   -33.924900,
   18.424101,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.johannesburg.3',
   'afs1-az3',
   -33.924900,
   18.424101,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.nigeria.1',
   'afs1-los1-az1',
   6.524400,
   3.379200,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.hongkong.1',
   'ape1-az1',
   22.319300,
   114.169403,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.hongkong.2',
   'ape1-az2',
   22.319300,
   114.169403,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.hongkong.3',
   'ape1-az3',
   22.319300,
   114.169403,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.tokyo.1',
   'apne1-az1',
   35.676201,
   139.650299,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.tokyo.2',
   'apne1-az2',
   35.676201,
   139.650299,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.tokyo.4',
   'apne1-az4',
   35.676201,
   139.650299,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.taipei.1',
   'apne1-tpe1-az1',
   25.033001,
   121.565399,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.seoul.1',
   'apne2-az1',
   37.566502,
   126.977997,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.seoul.2',
   'apne2-az2',
   37.566502,
   126.977997,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.seoul.3',
   'apne2-az3',
   37.566502,
   126.977997,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.seoul.4',
   'apne2-az4',
   37.566502,
   126.977997,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.osaka.1',
   'apne3-az1',
   34.693699,
   135.502304,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.osaka.2',
   'apne3-az2',
   34.693699,
   135.502304,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.osaka.3',
   'apne3-az3',
   34.693699,
   135.502304,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.mumbai.1',
   'aps1-az1',
   19.076000,
   72.877701,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.mumbai.2',
   'aps1-az2',
   19.076000,
   72.877701,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.mumbai.3',
   'aps1-az3',
   19.076000,
   72.877701,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.kolkata.1',
   'aps1-ccu1-az1',
   22.572599,
   88.363899,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.delhi.1',
   'aps1-del1-az1',
   28.704100,
   77.102501,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.hyderabad.1',
   'aps2-az1',
   17.385000,
   78.486702,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.hyderabad.2',
   'aps2-az2',
   17.385000,
   78.486702,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.hyderabad.3',
   'aps2-az3',
   17.385000,
   78.486702,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.singapore.1',
   'apse1-az1',
   1.352100,
   103.819801,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.singapore.2',
   'apse1-az2',
   1.352100,
   103.819801,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.singapore.3',
   'apse1-az3',
   1.352100,
   103.819801,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.bangkok.1',
   'apse1-bkk1-az1',
   13.756300,
   100.501801,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.sydney.1',
   'apse2-az1',
   -33.868801,
   151.209305,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.sydney.2',
   'apse2-az2',
   -33.868801,
   151.209305,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.sydney.3',
   'apse2-az3',
   -33.868801,
   151.209305,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.perth.1',
   'apse2-per1-az1',
   -31.952299,
   115.861298,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.jakarta.1',
   'apse3-az1',
   -6.208800,
   106.845596,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.jakarta.2',
   'apse3-az2',
   -6.208800,
   106.845596,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.jakarta.3',
   'apse3-az3',
   -6.208800,
   106.845596,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.melbourne.1',
   'apse4-az1',
   -37.813599,
   144.963104,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.melbourne.2',
   'apse4-az2',
   -37.813599,
   144.963104,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.melbourne.3',
   'apse4-az3',
   -37.813599,
   144.963104,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.montreal.1',
   'cac1-az1',
   45.501900,
   -73.567398,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.montreal.2',
   'cac1-az2',
   45.501900,
   -73.567398,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.montreal.4',
   'cac1-az4',
   45.501900,
   -73.567398,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.frankfurt.1',
   'euc1-az1',
   50.110901,
   8.682100,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.frankfurt.2',
   'euc1-az2',
   50.110901,
   8.682100,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.frankfurt.3',
   'euc1-az3',
   50.110901,
   8.682100,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.hamburg.1',
   'euc1-ham1-az1',
   53.548801,
   9.987200,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.warsaw.1',
   'euc1-waw1-az1',
   52.229698,
   21.012199,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.zurich.1',
   'euc2-az1',
   47.376900,
   8.541700,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.zurich.2',
   'euc2-az2',
   47.376900,
   8.541700,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.zurich.3',
   'euc2-az3',
   47.376900,
   8.541700,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.stockholm.1',
   'eun1-az1',
   59.329300,
   18.068600,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.stockholm.2',
   'eun1-az2',
   59.329300,
   18.068600,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.stockholm.3',
   'eun1-az3',
   59.329300,
   18.068600,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.copenhagen.1',
   'eun1-cph1-az1',
   55.676102,
   12.568300,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.finland.1',
   'eun1-hel1-az1',
   60.169899,
   24.938400,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.milan.1',
   'eus1-az1',
   45.464199,
   9.190000,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.milan.2',
   'eus1-az2',
   45.464199,
   9.190000,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.milan.3',
   'eus1-az3',
   45.464199,
   9.190000,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.spain.1',
   'eus2-az1',
   41.597599,
   -0.905700,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.spain.2',
   'eus2-az2',
   41.597599,
   -0.905700,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.spain.3',
   'eus2-az3',
   41.597599,
   -0.905700,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.ireland.1',
   'euw1-az1',
   53.779800,
   -7.305500,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.ireland.2',
   'euw1-az2',
   53.779800,
   -7.305500,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.ireland.3',
   'euw1-az3',
   53.779800,
   -7.305500,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.london.1',
   'euw2-az1',
   51.507198,
   -0.127600,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.london.2',
   'euw2-az2',
   51.507198,
   -0.127600,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.london.3',
   'euw2-az3',
   51.507198,
   -0.127600,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.paris.1',
   'euw3-az1',
   48.856602,
   2.352200,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.paris.2',
   'euw3-az2',
   48.856602,
   2.352200,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.paris.3',
   'euw3-az3',
   48.856602,
   2.352200,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.uae.1',
   'mec1-az1',
   23.424101,
   53.847801,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.uae.2',
   'mec1-az2',
   23.424101,
   53.847801,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.uae.3',
   'mec1-az3',
   23.424101,
   53.847801,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.bahrain.1',
   'mes1-az1',
   26.066700,
   50.557701,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.bahrain.2',
   'mes1-az2',
   26.066700,
   50.557701,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.bahrain.3',
   'mes1-az3',
   26.066700,
   50.557701,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.oman.1',
   'mes1-mct1-az1',
   23.587999,
   58.382900,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.saopaulo.1',
   'sae1-az1',
   -23.555799,
   -46.639599,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.saopaulo.2',
   'sae1-az2',
   -23.555799,
   -46.639599,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.saopaulo.3',
   'sae1-az3',
   -23.555799,
   -46.639599,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.atlanta.1',
   'use1-atl1-az1',
   33.748798,
   -84.387703,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.virginia.1',
   'use1-az1',
   39.043800,
   -77.487396,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.virginia.2',
   'use1-az2',
   39.043800,
   -77.487396,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.virginia.3',
   'use1-az3',
   39.043800,
   -77.487396,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.virginia.4',
   'use1-az4',
   39.043800,
   -77.487396,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.virginia.5',
   'use1-az5',
   39.043800,
   -77.487396,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.virginia.6',
   'use1-az6',
   39.043800,
   -77.487396,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.boston.1',
   'use1-bos1-az1',
   42.360100,
   -71.058899,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.buenosaires.1',
   'use1-bue1-az1',
   -34.603699,
   -58.381599,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.chicago.1',
   'use1-chi1-az1',
   41.878101,
   -87.629799,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.dallas.1',
   'use1-dfw1-az1',
   32.776699,
   -96.796997,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.houston.1',
   'use1-iah1-az1',
   29.760401,
   -95.369797,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.lima.1',
   'use1-lim1-az1',
   -12.046400,
   -77.042801,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.kansas.1',
   'use1-mci1-az1',
   39.099701,
   -94.578598,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.miami.1',
   'use1-mia1-az1',
   25.761700,
   -80.191803,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.minneapolis.1',
   'use1-msp1-az1',
   44.977798,
   -93.264999,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.newyork.1',
   'use1-nyc1-az1',
   40.712799,
   -74.005997,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.philadelphia.1',
   'use1-phl1-az1',
   39.952599,
   -75.165199,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.mexico.1',
   'use1-qro1-az1',
   23.634501,
   -102.552803,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.santiago.1',
   'use1-scl1-az1',
   -33.448898,
   -70.669296,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.ohio.1',
   'use2-az1',
   40.417301,
   -82.907097,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.ohio.2',
   'use2-az2',
   40.417301,
   -82.907097,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.ohio.3',
   'use2-az3',
   40.417301,
   -82.907097,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.sanjose.1',
   'usw1-az1',
   37.338699,
   -121.885300,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.sanjose.3',
   'usw1-az3',
   37.338699,
   -121.885300,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.oregon.1',
   'usw2-az1',
   45.839901,
   -119.700600,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.oregon.2',
   'usw2-az2',
   45.839901,
   -119.700600,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.oregon.3',
   'usw2-az3',
   45.839901,
   -119.700600,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.oregon.4',
   'usw2-az4',
   45.839901,
   -119.700600,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.denver.1',
   'usw2-den1-az1',
   39.739201,
   -104.990303,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.lasvegas.1',
   'usw2-las1-az1',
   36.171600,
   -115.139099,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.losangeles.1',
   'usw2-lax1-az1',
   34.052200,
   -118.243698,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.losangeles.2',
   'usw2-lax1-az2',
   34.052200,
   -118.243698,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.portland.1',
   'usw2-pdx1-az1',
   45.515202,
   -122.678398,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.phoenix.1',
   'usw2-phx1-az1',
   33.448399,
   -112.073997,
   (select seller_id from sellers where seller_name = 'amazon')
);

INSERT INTO datacenters(
	datacenter_name,
	native_name,
	latitude,
	longitude,
	seller_id)
VALUES(
   'amazon.seattle.1',
   'usw2-sea1-az1',
   47.606201,
   -122.332100,
   (select seller_id from sellers where seller_name = 'amazon')
);
