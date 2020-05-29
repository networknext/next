SELECT
  buyerId,
  date,
  ROUND((SUM(sliceBillNibblins) / 1000000000) / 100, 2) AS sliceBillUsd,
  ROUND(SUM(bandwidthBytes) / 1000000000) AS totalBandwidthGb,
  ROUND((AVG(avgNibblinsPerGb) / 1000000000) / 100, 4) AS avgUsdPerGb,
  ROUND(AVG(avgImprovementMs), 2) AS avgImprovementMs,
  SUM(IF(onNetworkNext, 1, 0)) AS sessionsOnNetworkNext,
  ROUND(SUM(IF(onNetworkNext, 1, 0)) / COUNT(sessionId) * 100, 2) AS percentageOnNetworkNext,
  FLOOR(SUM(secondsWith20msOrGreater) / 60 / 60) AS hoursWith20msOrGreater,
  FLOOR(SUM(secondsOnNetworkNext) / 60 / 60) AS hoursOnNetworkNext,
  FLOOR(SUM(secondsImproved) / 60 / 60) AS hoursImproved,
  FLOOR(SUM(secondsDegraded) / 60 / 60) AS hoursDegraded,
  FLOOR(SUM(secondsNoMeasurement) / 60 / 60) AS hoursNoMeasurement,
  FLOOR(SUM(seconds0To5Ms) / 60 / 60) AS hours0To5Ms,
  FLOOR(SUM(seconds5To10Ms) / 60 / 60) AS hours5To10Ms,
  FLOOR(SUM(seconds10To15Ms) / 60 / 60) AS hours10To15Ms,
  FLOOR(SUM(seconds15To20Ms) / 60 / 60) AS hours15To20Ms,
  FLOOR(SUM(seconds20To30Ms) / 60 / 60) AS hours20To30Ms,
  FLOOR(SUM(seconds30To50Ms) / 60 / 60) AS hours30To50Ms,
  FLOOR(SUM(seconds50To100Ms) / 60 / 60) AS hours50To100Ms,
  FLOOR(SUM(seconds100PlusMs) / 60 / 60) AS hours100PlusMs,
  FLOOR(SUM(secondsPacketLossReduced) / 60 / 60) AS hoursPacketLossReduced
FROM (
  SELECT
    buyerId,
    sessionId,
    DATE(timestampStart) AS date,
    SUM(sliceBill) AS sliceBillNibblins,
    SUM(bandwidthBytes) AS bandwidthBytes,
    IF(SUM(bandwidthBytes) = 0, NULL, (SUM(sliceBill) / SUM(bandwidthBytes)) * 1000000000) AS avgNibblinsPerGb,
    AVG(improvementMs) AS avgImprovementMs,
    MAX(onNetworkNext) AS onNetworkNext,
    SUM(secondsWith20msOrGreater) AS secondsWith20msOrGreater,
    SUM(secondsOnNetworkNext) AS secondsOnNetworkNext,
    SUM(secondsImproved) AS secondsImproved,
    SUM(secondsDegraded) AS secondsDegraded,
    SUM(secondsNoMeasurement) AS secondsNoMeasurement,
    SUM(seconds0To5Ms) AS seconds0To5Ms,
    SUM(seconds5To10Ms) AS seconds5To10Ms,
    SUM(seconds10To15Ms) AS seconds10To15Ms,
    SUM(seconds15To20Ms) AS seconds15To20Ms,
    SUM(seconds20To30Ms) AS seconds20To30Ms,
    SUM(seconds30To50Ms) AS seconds30To50Ms,
    SUM(seconds50To100Ms) AS seconds50To100Ms,
    SUM(seconds100PlusMs) AS seconds100PlusMs,
    SUM(secondsPacketLossReduced) AS secondsPacketLossReduced,
  FROM (
    SELECT 
      buyerId,
      sessionId,
      timestampStart,
      timestamp,
      IF(wasRouteIssued, (
        LEAST(
          (
            SELECT
              COALESCE(SUM(route.priceIngress) + SUM(route.priceEgress), 0) /* seller charge */
            FROM UNNEST(route) AS route
          )
          + (4000000000 / 1000000000) * (bytesBandwidth) /* Network Next cut */
          , (25000000000 / 1000000000) * (bytesBandwidth) /* 25c maximum */
        )
      ), 0) AS sliceBill,
      fallbackToDirect AS fallbackToDirect,
      tryBeforeYouBuy AS tryBeforeYouBuy,
      IF(wasRouteIssued, bytesBandwidth, 0) AS bandwidthBytes,
      IF(onNetworkNext AND GREATEST(0, improvementRtt) > 0, GREATEST(0, improvementRtt), NULL) AS improvementMs,
      onNetworkNext,
      IF(onNetworkNext AND improvementRtt > 0 AND improvementRtt < 5, 1, 0) * 10 AS seconds0To5Ms,
      IF(onNetworkNext AND improvementRtt >= 5 AND improvementRtt < 10, 1, 0) * 10 AS seconds5To10Ms,
      IF(onNetworkNext AND improvementRtt >= 10 AND improvementRtt < 15, 1, 0) * 10 AS seconds10To15Ms,
      IF(onNetworkNext AND improvementRtt >= 15 AND improvementRtt < 20, 1, 0) * 10 AS seconds15To20Ms,
      IF(onNetworkNext AND improvementRtt >= 20 AND improvementRtt < 30, 1, 0) * 10 AS seconds20To30Ms,
      IF(onNetworkNext AND improvementRtt >= 30 AND improvementRtt < 50, 1, 0) * 10 AS seconds30To50Ms,
      IF(onNetworkNext AND improvementRtt >= 50 AND improvementRtt < 100, 1, 0) * 10 AS seconds50To100Ms,
      IF(onNetworkNext AND improvementRtt >= 100, 1, 0) * 10 AS seconds100PlusMs,
      IF(onNetworkNext AND improvementRtt >= 20, 1, 0) * 10 AS secondsWith20msOrGreater,
      IF(onNetworkNext, 1, 0) * 10 AS secondsOnNetworkNext,
      IF(onNetworkNext AND improvementRtt > 0, 1, 0) * 10 AS secondsImproved,
      IF(onNetworkNext AND improvementRtt < 0, 1, 0) * 10 AS secondsDegraded,
      IF(onNetworkNext AND (improvementRtt = 0 OR improvementRtt IS NULL), 1, 0) * 10 AS secondsNoMeasurement,
      IF(onNetworkNext AND improvementPacketLoss > 0, 1, 0) * 10 AS secondsPacketLossReduced,
    FROM (
      SELECT
        *
      FROM (
        SELECT
          buyerId,
          a.sessionId,
          timestampStart,
          timestamp,
          route,
          fallbackToDirect,
          tryBeforeYouBuy,
          networkNext,
          bytesUp + bytesDown AS bytesBandwidth,
          LEAD(nextRtt, 1) OVER (PARTITION BY a.sessionId, timestampStart ORDER BY timestamp ASC) AS nextRtt,
          LEAD(directRtt, 1) OVER (PARTITION BY a.sessionId, timestampStart ORDER BY timestamp ASC) AS directRtt,
          LEAD(directRtt, 1) OVER (PARTITION BY a.sessionId, timestampStart ORDER BY timestamp ASC) - LEAD(nextRtt, 1) OVER (PARTITION BY a.sessionId, timestampStart ORDER BY timestamp ASC) AS improvementRtt,
          LEAD(nextPacketLoss, 1) OVER (PARTITION BY a.sessionId, timestampStart ORDER BY timestamp ASC) AS nextPacketLoss,
          LEAD(directPacketLoss, 1) OVER (PARTITION BY a.sessionId, timestampStart ORDER BY timestamp ASC) AS directPacketLoss,
          LEAD(directPacketLoss, 1) OVER (PARTITION BY a.sessionId, timestampStart ORDER BY timestamp ASC) - LEAD(nextPacketLoss, 1) OVER (PARTITION BY a.sessionId, timestampStart ORDER BY timestamp ASC) AS improvementPacketLoss,
          bytesUp,
          bytesDown,
          (networkNext AND nextRtt > 0 AND nextRtt IS NOT NULL) AS onNetworkNext,
          (networkNext AND ARRAY_LENGTH(route) > 0) AS wasRouteIssued,
          COALESCE(maxImprovementRtt, 0) AS maxImprovementRtt,
          COALESCE(slicesNetImprovedPacketLoss, 0) AS slicesNetImprovedPacketLoss,          
        FROM `network-next-v3-prod.v3.billing` AS a
        LEFT JOIN (
          SELECT
            sessionId,
            MAX(improvementRtt) AS maxImprovementRtt,
            COUNTIF(improvementPacketLoss > 0) - COUNTIF(improvementPacketLoss < 0) AS slicesNetImprovedPacketLoss,
          FROM (
            SELECT
              *
            FROM (
              SELECT
                sessionId,
                LEAD(nextRtt, 1) OVER (PARTITION BY sessionId, timestampStart ORDER BY timestamp ASC) AS nextRtt,
                LEAD(directRtt, 1) OVER (PARTITION BY sessionId, timestampStart ORDER BY timestamp ASC) AS directRtt,
                LEAD(directRtt, 1) OVER (PARTITION BY sessionId, timestampStart ORDER BY timestamp ASC) - LEAD(nextRtt, 1) OVER (PARTITION BY sessionId, timestampStart ORDER BY timestamp ASC) AS improvementRtt,
                LEAD(nextPacketLoss, 1) OVER (PARTITION BY sessionId, timestampStart ORDER BY timestamp ASC) AS nextPacketLoss,
                LEAD(directPacketLoss, 1) OVER (PARTITION BY sessionId, timestampStart ORDER BY timestamp ASC) AS directPacketLoss,
                LEAD(directPacketLoss, 1) OVER (PARTITION BY sessionId, timestampStart ORDER BY timestamp ASC) - LEAD(nextPacketLoss, 1) OVER (PARTITION BY sessionId, timestampStart ORDER BY timestamp ASC) AS improvementPacketLoss,
              FROM `network-next-v3-prod.v3.billing`
              WHERE DATE(timestampStart) >= @start AND DATE(timestampStart) <= @end
            ) 
            WHERE nextRtt > 0
          )
          GROUP BY sessionId
        ) AS b
          ON b.sessionId = a.sessionId
      )
      WHERE (NOT buyerId = "Buyer/dtqb5J5pAGMS7IizFfWD") OR (maxImprovementRtt >= 20 OR slicesNetImprovedPacketLoss >= 3)
    )
    WHERE DATE(timestampStart) >= @start AND DATE(timestampStart) <= @end
  )
  GROUP BY buyerId, sessionId, DATE(timestampStart)
)
GROUP BY buyerId, date
ORDER BY date ASC