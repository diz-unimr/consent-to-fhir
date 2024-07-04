USE `notification_service`;
UPDATE `configuration` SET `value` = 
'<org.emau.icmvc.ttp.notification.service.model.config.NotificationConfig>
    <consumerConfigs>
        <param>
            <key>HTTPConsumer</key>
            <value class="org.emau.icmvc.ttp.notification.service.model.config.ConsumerConfig">
                <connectionType>HTTP</connectionType>
                <messageTypes>
                    <string>*</string>
                </messageTypes>
                <excludeClientIdFilter
                    class="set">
                    <string>E-PIX_Web</string>
                </excludeClientIdFilter>
                <parameter>
                    <param>
                        <key>url</key>
                        <value>http://host.docker.internal:8080/notification</value>
                    </param>
                    <param>
                        <key>username</key>
                        <value>test</value>
                    </param>
                    <param>
                        <key>password</key>
                        <value>test</value>
                    </param>
                </parameter>
            </value>
        </param>
    </consumerConfigs>
</org.emau.icmvc.ttp.notification.service.model.config.NotificationConfig>'
WHERE `configKey` = 'notification.config';
