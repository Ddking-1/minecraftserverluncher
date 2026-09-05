// SPDX-License-Identifier: GPL-3.0-only
#include "BaseInstance.h"

BaseInstance::BaseInstance(QObject *parent) : QObject(parent)
{
}

BaseInstance::~BaseInstance() = default;

QString BaseInstance::id() const
{
    return m_id;
}

void BaseInstance::setId(const QString &id)
{
    if (m_id == id)
        return;
    m_id = id;
    emit propertiesChanged(this);
}

QString BaseInstance::instanceRoot() const
{
    return m_rootDir;
}

void BaseInstance::setInstanceRoot(const QString &root)
{
    if (m_rootDir == root)
        return;
    m_rootDir = root;
    emit propertiesChanged(this);
}

QString BaseInstance::name() const
{
    return m_name;
}

void BaseInstance::setName(const QString &name)
{
    if (m_name == name)
        return;
    m_name = name;
    emit propertiesChanged(this);
}

bool BaseInstance::isRunning() const
{
    return m_isRunning;
}

void BaseInstance::setRunning(bool running)
{
    if (m_isRunning == running)
        return;
    m_isRunning = running;
    emit runningStatusChanged(running);
}
