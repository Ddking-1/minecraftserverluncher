// SPDX-License-Identifier: GPL-3.0-only
#include "EditorConfig.h"

#include "InstancePersistence.h"

EditorConfig EditorConfig::defaults()
{
    EditorConfig cfg;
    cfg.title = QStringLiteral("Minecraft Server");
    cfg.backgroundImage.clear();
    cfg.accentColor = QStringLiteral("#3FB950");
    cfg.showNewsTab = true;
    cfg.showSettingsTab = true;
    return cfg;
}

QJsonObject EditorConfig::toJson() const
{
    QJsonObject obj;
    obj.insert(QStringLiteral("title"), title);
    obj.insert(QStringLiteral("background_image"), backgroundImage);
    obj.insert(QStringLiteral("accent_color"), accentColor);
    obj.insert(QStringLiteral("show_news_tab"), showNewsTab);
    obj.insert(QStringLiteral("show_settings_tab"), showSettingsTab);
    return obj;
}

EditorConfig EditorConfig::fromJson(const QJsonObject &obj)
{
    EditorConfig cfg = defaults();
    if (obj.contains(QStringLiteral("title")))
        cfg.title = obj.value(QStringLiteral("title")).toString();
    if (obj.contains(QStringLiteral("background_image")))
        cfg.backgroundImage = obj.value(QStringLiteral("background_image")).toString();
    if (obj.contains(QStringLiteral("accent_color")))
        cfg.accentColor = obj.value(QStringLiteral("accent_color")).toString();
    if (obj.contains(QStringLiteral("show_news_tab")))
        cfg.showNewsTab = obj.value(QStringLiteral("show_news_tab")).toBool(true);
    if (obj.contains(QStringLiteral("show_settings_tab")))
        cfg.showSettingsTab = obj.value(QStringLiteral("show_settings_tab")).toBool(true);
    return cfg;
}

bool EditorConfig::save(const QString &path) const
{
    return InstancePersistence::writeJsonFile(path, toJson());
}

bool EditorConfig::load(const QString &path, EditorConfig &out)
{
    QJsonObject obj;
    if (!InstancePersistence::readJsonFile(path, obj))
        return false;
    out = fromJson(obj);
    return true;
}
