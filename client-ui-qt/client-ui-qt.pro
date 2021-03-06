#-------------------------------------------------
#
# Project created by QtCreator 2017-06-04T09:41:45
#
#-------------------------------------------------

QT       += core gui network charts

greaterThan(QT_MAJOR_VERSION, 4): QT += widgets

TARGET = Shitama
TEMPLATE = app

# The following define makes your compiler emit warnings if you use
# any feature of Qt which as been marked as deprecated (the exact warnings
# depend on your compiler). Please consult the documentation of the
# deprecated API in order to know how to port your code away from it.
DEFINES += QT_DEPRECATED_WARNINGS

DEFINES += SHITAMA_BUILD_ID=\\\"$$(SHITAMA_BUILD_ID)\\\" SHITAMA_COMMIT=\\\"$$(SHITAMA_COMMIT)\\\"

# You can also make your code fail to compile if you use deprecated APIs.
# In order to do so, uncomment the following line.
# You can also select to disable deprecated APIs only up to a certain version of Qt.
#DEFINES += QT_DISABLE_DEPRECATED_BEFORE=0x060000    # disables all the APIs deprecated before Qt 6.0.0

SOURCES += \
        main.cpp \
        mainwindow.cpp \
    api.cpp \
    connectiondialog.cpp

HEADERS += \
        mainwindow.h \
    api.h \
    connectiondialog.h

FORMS += \
        mainwindow.ui \
    connectiondialog.ui

RC_FILE = shitama.rc

DESTDIR = ../build/client-ui-qt/

DISTFILES += \
    shitama.rc \
    shitama.ico

RESOURCES += \
    shitama.qrc
